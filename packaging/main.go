package main

import (
	"archive/tar"
	"archive/zip"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ulikunitz/xz"
)

//go:embed postgres-sources.json licenses runtime-notices.txt
var assets embed.FS

type artifact struct {
	Artifact string `json:"artifact"`
	Archive  string `json:"archive"`
	SHA256   string `json:"sha256"`
}
type vcRuntime struct {
	Version        string   `json:"version"`
	URL            string   `json:"url"`
	SHA256         string   `json:"sha256"`
	ChecksumSource string   `json:"checksum_source"`
	Files          []string `json:"files"`
}
type sources struct {
	PostgresVersion  string              `json:"postgres_version"`
	PublisherVersion string              `json:"publisher_version"`
	Publisher        string              `json:"publisher"`
	Repository       string              `json:"repository"`
	Targets          map[string]artifact `json:"targets"`
	WindowsVCRuntime vcRuntime           `json:"windows_vc_runtime"`
}
type manifest struct {
	FormatVersion   int               `json:"format_version"`
	Target          string            `json:"target"`
	SourceRevision  string            `json:"source_revision"`
	SourceDirty     bool              `json:"source_dirty"`
	PostgresVersion string            `json:"postgres_version"`
	PostgresURL     string            `json:"postgres_url"`
	PostgresSHA256  string            `json:"postgres_sha256"`
	Publisher       string            `json:"publisher"`
	RuntimeOnly     bool              `json:"runtime_only"`
	Files           map[string]string `json:"files_sha256"`
}

func main() {
	target := flag.String("target", runtime.GOOS+"-"+runtime.GOARCH, "windows-amd64, linux-amd64, linux-arm64, darwin-amd64, darwin-arm64")
	out := flag.String("out", "", "new output folder beneath this repository's dist directory")
	runtimeOnly := flag.Bool("runtime-only", false, "prepare runtime, skills and licenses without building the CLI")
	offline := flag.Bool("offline", false, "use only the checksum-verified download cache")
	flag.Parse()
	if err := run(*target, *out, *runtimeOnly, *offline); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(target, out string, runtimeOnly, offline bool) error {
	var pins sources
	raw, err := assets.ReadFile("postgres-sources.json")
	if err != nil {
		return err
	}
	if err = json.Unmarshal(raw, &pins); err != nil {
		return err
	}
	pin, ok := pins.Targets[target]
	if !ok {
		return fmt.Errorf("unsupported target %q", target)
	}
	repo, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, err = os.Stat(filepath.Join(repo, "go.mod")); err != nil {
		return errors.New("run packaging from the repository root")
	}
	dist := filepath.Join(repo, "dist")
	if out == "" {
		out = filepath.Join(dist, "scopeweaver-cli-"+target)
	}
	out, err = filepath.Abs(out)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(dist, out)
	if err != nil || !filepath.IsLocal(rel) || rel == "." {
		return errors.New("output must be a new folder beneath dist")
	}
	if err = plainParents(out); err != nil {
		return err
	}
	cache := filepath.Join(dist, "cache")
	if err = plainParents(cache); err != nil {
		return err
	}
	if err = os.MkdirAll(cache, 0755); err != nil {
		return err
	}
	filename := "embedded-postgres-binaries-" + pin.Artifact + "-" + pins.PublisherVersion + ".jar"
	url := pins.Repository + "embedded-postgres-binaries-" + pin.Artifact + "/" + pins.PublisherVersion + "/" + filename
	cached := filepath.Join(cache, filename)
	if err = download(cached, url, pin.SHA256, offline); err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	if err = os.Mkdir(out, 0755); err != nil {
		return fmt.Errorf("output must not exist: %w", err)
	}
	root, err := os.OpenRoot(out)
	if err != nil {
		return err
	}
	defer root.Close()
	if err = extractJar(cached, pin.Archive, root); err != nil {
		return err
	}
	if target == "windows-amd64" {
		runtimeFile := filepath.Join(cache, "Microsoft.VCLibs.x64.14.00.Desktop.appx")
		if err = download(runtimeFile, pins.WindowsVCRuntime.URL, pins.WindowsVCRuntime.SHA256, offline); err != nil {
			return err
		}
		if err = extractVCRuntime(runtimeFile, pins.WindowsVCRuntime, root); err != nil {
			return err
		}
	}

	if err = copyFile(root, "LICENSE", filepath.Join(repo, "LICENSE"), 0644); err != nil {
		return err
	}
	originalNotices, err := os.ReadFile(filepath.Join(repo, "THIRD_PARTY_NOTICES.txt"))
	if err != nil {
		return err
	}
	runtimeNotices, err := assets.ReadFile("runtime-notices.txt")
	if err != nil {
		return err
	}
	if err = writeFile(root, "THIRD_PARTY_NOTICES.txt", append(originalNotices, runtimeNotices...), 0644); err != nil {
		return err
	}
	for _, doc := range []string{"README.md", "README.ko.md", "docs/cli.md", "docs/ko/cli.md", "docs/CLI-VERIFICATION.md", "docs/ko/CLI-VERIFICATION.md"} {
		if err = copyFile(root, doc, filepath.Join(repo, filepath.FromSlash(doc)), 0644); err != nil {
			return err
		}
	}
	if err = copyTree(root, "skills", filepath.Join(repo, "skills")); err != nil {
		return err
	}
	if err = fs.WalkDir(assets, "licenses", func(name string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		b, e := assets.ReadFile(name)
		if e != nil {
			return e
		}
		return writeFile(root, name, b, 0644)
	}); err != nil {
		return err
	}
	if err = writeFile(root, "POSTGRES-SOURCES.json", raw, 0644); err != nil {
		return err
	}
	if !runtimeOnly {
		parts := strings.Split(target, "-")
		exe := "scopeweaver-cli"
		if parts[0] == "windows" {
			exe += ".exe"
		}
		cmd := exec.Command("go", "build", "-trimpath", "-o", filepath.Join(out, exe), "./cmd/scopeweaver-cli")
		cmd.Env = buildEnv(parts[0], parts[1])
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("build CLI: %w", err)
		}
		if err = copyGoLicenses(root, parts[0], parts[1]); err != nil {
			return err
		}
	}
	rev := exec.Command("git", "rev-parse", "HEAD")
	revision, err := rev.Output()
	if err != nil {
		return err
	}
	status, err := exec.Command("git", "status", "--porcelain", "--untracked-files=normal").Output()
	if err != nil {
		return err
	}
	m := manifest{1, target, strings.TrimSpace(string(revision)), len(status) > 0, pins.PostgresVersion, url, pin.SHA256, pins.Publisher, runtimeOnly, map[string]string{}}
	if err = filepath.WalkDir(out, func(name string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.Type().IsRegular() {
			return nil
		}
		sum, e := fileHash(name)
		if e != nil {
			return e
		}
		r, e := filepath.Rel(out, name)
		if e != nil {
			return e
		}
		m.Files[filepath.ToSlash(r)] = sum
		return nil
	}); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err = writeFile(root, "MANIFEST.json", append(encoded, '\n'), 0644); err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}

func buildEnv(goos, goarch string) []string {
	var result []string
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "GOOS=") || strings.HasPrefix(v, "GOARCH=") || strings.HasPrefix(v, "CGO_ENABLED=") {
			continue
		}
		result = append(result, v)
	}
	return append(result, "GOOS="+goos, "GOARCH="+goarch, "CGO_ENABLED=0")
}

func plainParents(name string) error {
	for p := filepath.Clean(name); ; p = filepath.Dir(p) {
		info, err := os.Lstat(p)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink parent is not allowed: %s", p)
		}
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if filepath.Dir(p) == p {
			return nil
		}
	}
}

func fileHash(name string) (string, error) {
	f, e := os.Open(name)
	if e != nil {
		return "", e
	}
	defer f.Close()
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		return "", e
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func verify(name, want string) error {
	got, e := fileHash(name)
	if e != nil {
		return e
	}
	if got != want {
		return fmt.Errorf("SHA256 mismatch for %s: got %s, want %s", name, got, want)
	}
	return nil
}

func download(name, url, sum string, offline bool) error {
	if _, err := os.Lstat(name); err == nil {
		return verify(name, sum)
	} else if !os.IsNotExist(err) {
		return err
	}
	if offline {
		return fmt.Errorf("offline cache missing: %s", name)
	}
	client := http.Client{Timeout: 5 * time.Minute}
	response, err := client.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: HTTP %s", url, response.Status)
	}
	f, err := os.CreateTemp(filepath.Dir(name), ".download-*")
	if err != nil {
		return err
	}
	// Keep a failed download for inspection; never overwrite an existing cache entry.
	if _, err = io.Copy(f, io.LimitReader(response.Body, 512<<20)); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = verify(f.Name(), sum); err != nil {
		return err
	}
	if _, err = os.Lstat(name); err == nil {
		return errors.New("cache appeared while downloading")
	}
	return os.Rename(f.Name(), name)
}

func safeName(name string) (string, error) {
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

func extractJar(name, member string, root *os.Root) error {
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
			return extractTar(tar.NewReader(dec), root)
		}
	}
	return fmt.Errorf("missing archive %s", member)
}

func extractTar(t *tar.Reader, root *os.Root) error {
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
		name, err := safeName(h.Name)
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
			target, err := safeName(path.Join(path.Dir(strings.TrimPrefix(name, "runtime/postgres/")), h.Linkname))
			if err != nil {
				return err
			}
			if path.IsAbs(h.Linkname) || strings.ContainsAny(h.Linkname, "\\:\x00") {
				return errors.New("unsafe symlink target")
			}
			links = append(links, archiveLink{name, "runtime/postgres/" + target})
		case tar.TypeLink:
			target, err := safeName(h.Linkname)
			if err != nil {
				return err
			}
			links = append(links, archiveLink{name, "runtime/postgres/" + target})
		default:
			return fmt.Errorf("unsupported archive entry type %d for %s", h.Typeflag, h.Name)
		}
	}
}

func writeFile(root *os.Root, name string, data []byte, mode fs.FileMode) error {
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
func copyFile(root *os.Root, name, source string, mode fs.FileMode) error {
	b, e := os.ReadFile(source)
	if e != nil {
		return e
	}
	return writeFile(root, name, b, mode)
}
func copyTree(root *os.Root, dest, source string) error {
	return filepath.WalkDir(source, func(name string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("nonregular source: %s", name)
		}
		r, e := filepath.Rel(source, name)
		if e != nil {
			return e
		}
		return copyFile(root, path.Join(dest, filepath.ToSlash(r)), name, 0644)
	})
}

func copyGoLicenses(root *os.Root, goos, goarch string) error {
	cmd := exec.Command("go", "list", "-deps", "-f", "{{with .Module}}{{.Path}}|{{.Version}}|{{.Dir}}{{end}}", "./cmd/scopeweaver-cli")
	cmd.Env = buildEnv(goos, goarch)
	raw, err := cmd.Output()
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		p := strings.Split(strings.TrimSpace(line), "|")
		if len(p) != 3 || p[1] == "" || seen[p[0]] {
			continue
		}
		seen[p[0]] = true
		entries, err := os.ReadDir(p[2])
		if err != nil {
			return err
		}
		for _, entry := range entries {
			n := strings.ToLower(entry.Name())
			if entry.IsDir() || !(strings.HasPrefix(n, "license") || strings.HasPrefix(n, "copying") || strings.HasPrefix(n, "notice")) {
				continue
			}
			if err = copyFile(root, path.Join("licenses/go", p[0]+"@"+p[1], entry.Name()), filepath.Join(p[2], entry.Name()), 0644); err != nil {
				return err
			}
		}
	}
	// The standard library's license accompanies every Go-built binary.
	return copyFile(root, "licenses/go/Go-LICENSE", filepath.Join(runtime.GOROOT(), "LICENSE"), 0644)
}

func extractVCRuntime(name string, pin vcRuntime, root *os.Root) error {
	z, err := zip.OpenReader(name)
	if err != nil {
		return err
	}
	defer z.Close()
	for _, member := range pin.Files {
		found := false
		for _, f := range z.File {
			if f.Name == member {
				if _, err = safeName(member); err != nil {
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
				if err = writeFile(root, "runtime/postgres/bin/"+member, data, 0644); err != nil {
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
			if err = writeFile(root, link.name, data, info.Mode().Perm()); err != nil {
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
