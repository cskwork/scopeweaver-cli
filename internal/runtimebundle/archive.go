package runtimebundle

import (
	"archive/tar"
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

func SafeName(name string) (string, error) {
	if strings.ContainsAny(name, "\\:\x00") || strings.HasPrefix(name, "/") {
		return "", fmt.Errorf("unsafe archive path %q", name)
	}
	for _, part := range strings.Split(name, "/") {
		if part == "." || part == ".." {
			continue
		}
		base := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		reserved := base == "CON" || base == "PRN" || base == "AUX" || base == "NUL"
		reserved = reserved || len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9'
		if reserved || strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
			return "", fmt.Errorf("unsafe portable archive path %q", name)
		}
	}
	cleaned := path.Clean(name)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || !filepath.IsLocal(filepath.FromSlash(cleaned)) {
		return "", fmt.Errorf("unsafe archive path %q", name)
	}
	return cleaned, nil
}

func ExtractJar(name, member string, root *os.Root) error {
	z, err := zip.OpenReader(name)
	if err != nil {
		return err
	}
	defer z.Close()
	for _, f := range z.File {
		if f.Name == member {
			r, err := f.Open()
			if err != nil {
				return err
			}
			defer r.Close()
			dec, err := xz.NewReader(r)
			if err != nil {
				return err
			}
			return ExtractTar(tar.NewReader(dec), root)
		}
	}
	return fmt.Errorf("missing archive %s", member)
}

func ExtractTar(t *tar.Reader, root *os.Root) error {
	var links []archiveLink
	var total int64
	for {
		h, err := t.Next()
		if err == io.EOF {
			return materializeLinks(links, root)
		}
		if err != nil {
			return err
		}
		name, err := SafeName(h.Name)
		if err != nil {
			return err
		}
		name = "runtime/postgres/" + name
		if err = root.MkdirAll(path.Dir(name), 0755); err != nil {
			return err
		}
		switch h.Typeflag {
		case tar.TypeDir:
			if err = root.MkdirAll(name, 0755); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			total += h.Size
			if h.Size < 0 || total > 2<<30 {
				return errors.New("PostgreSQL archive exceeds extraction limit")
			}
			f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, fs.FileMode(h.Mode)&0777)
			if err != nil {
				return err
			}
			_, err = io.CopyN(f, t, h.Size)
			closeErr := f.Close()
			if err != nil {
				return err
			}
			if closeErr != nil {
				return closeErr
			}
		case tar.TypeSymlink:
			// Relative library links must stay inside the private PostgreSQL tree.
			target, err := SafeName(path.Join(path.Dir(strings.TrimPrefix(name, "runtime/postgres/")), h.Linkname))
			if err != nil {
				return err
			}
			if path.IsAbs(h.Linkname) || strings.ContainsAny(h.Linkname, "\\:\x00") {
				return errors.New("unsafe symlink target")
			}
			links = append(links, archiveLink{name, "runtime/postgres/" + target})
		case tar.TypeLink:
			target, err := SafeName(h.Linkname)
			if err != nil {
				return err
			}
			links = append(links, archiveLink{name, "runtime/postgres/" + target})
		default:
			return fmt.Errorf("unsupported archive entry type %d for %s", h.Typeflag, h.Name)
		}
	}
}

func WriteFile(root *os.Root, name string, data []byte, mode fs.FileMode) error {
	if err := root.MkdirAll(path.Dir(name), 0755); err != nil {
		return err
	}
	f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func ExtractVCRuntime(name string, pin VCRuntime, root *os.Root) error {
	z, err := zip.OpenReader(name)
	if err != nil {
		return err
	}
	defer z.Close()
	for _, member := range pin.Files {
		found := false
		for _, f := range z.File {
			if f.Name == member {
				if _, err = SafeName(member); err != nil {
					return err
				}
				if f.UncompressedSize64 > 8<<20 {
					return fmt.Errorf("Visual C++ runtime DLL exceeds extraction limit: %s", member)
				}
				r, err := f.Open()
				if err != nil {
					return err
				}
				data, err := io.ReadAll(r)
				r.Close()
				if err != nil {
					return err
				}
				if err = WriteFile(root, "runtime/postgres/bin/"+member, data, 0644); err != nil {
					return err
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("missing Visual C++ runtime DLL %s", member)
		}
	}
	return nil
}

type archiveLink struct{ name, target string }

// Materialize internal library links as copies so Windows builders need no symlink privilege.
func materializeLinks(pending []archiveLink, root *os.Root) error {
	for len(pending) > 0 {
		remaining := make([]archiveLink, 0, len(pending))
		for _, link := range pending {
			source, err := root.Open(link.target)
			if os.IsNotExist(err) {
				remaining = append(remaining, link)
				continue
			}
			if err != nil {
				return err
			}
			info, err := source.Stat()
			if err != nil {
				source.Close()
				return err
			}
			if !info.Mode().IsRegular() {
				source.Close()
				return fmt.Errorf("archive link target is not a file: %s", link.target)
			}
			if info.Size() > 256<<20 {
				source.Close()
				return fmt.Errorf("archive link target exceeds copy limit: %s", link.target)
			}
			data, err := io.ReadAll(source)
			source.Close()
			if err != nil {
				return err
			}
			if err = WriteFile(root, link.name, data, info.Mode().Perm()); err != nil {
				return err
			}
		}
		if len(remaining) == len(pending) {
			return errors.New("archive has missing or circular link targets")
		}
		pending = remaining
	}
	return nil
}
