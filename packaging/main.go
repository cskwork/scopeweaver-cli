package main

import (
	"archive/tar"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cskwork/scopeweaver-cli/internal/runtimebundle"
)

//go:embed licenses runtime-notices.txt
var assets embed.FS

type sources = runtimebundle.Sources
type manifest struct {
	FormatVersion   int               `json:"format_version"`
	Target          string            `json:"target"`
	SourceRevision  string            `json:"source_revision"`
	SourceDirty     bool              `json:"source_dirty"`
	PostgresVersion string            `json:"postgres_version"`
	PostgresURL     string            `json:"postgres_url"`
	PostgresSHA256  string            `json:"postgres_sha256"`
	Publisher       string            `json:"publisher"`
	RuntimeMode     string            `json:"runtime_mode"`
	RuntimeOnly     bool              `json:"runtime_only"`
	Files           map[string]string `json:"files_sha256"`
}

func main() {
	target := flag.String("target", runtime.GOOS+"-"+runtime.GOARCH, "windows-amd64, linux-amd64, linux-arm64, darwin-amd64, darwin-arm64")
	out := flag.String("out", "", "new output folder beneath this repository's dist directory")
	bundleRuntime := flag.Bool("bundle-runtime", false, "bundle pinned runtime for local verification (not public download-mode packaging)")
	runtimeOnly := flag.Bool("runtime-only", false, "prepare runtime, skills and licenses without building the CLI")
	offline := flag.Bool("offline", false, "use only the checksum-verified download cache")
	flag.Parse()
	if err := runWithMode(*target, *out, *runtimeOnly, *offline, *bundleRuntime || *runtimeOnly); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(target, out string, runtimeOnly, offline bool) error {
	return runWithMode(target, out, runtimeOnly, offline, runtimeOnly)
}
func runWithMode(target, out string, runtimeOnly, offline, bundleRuntime bool) error {
	var pins sources
	raw := runtimebundle.SourceBytes()
	if err := json.Unmarshal(raw, &pins); err != nil {
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
	filename, url, err := pins.PostgresArtifact(target)
	if err != nil {
		return err
	}
	cache := filepath.Join(dist, "cache")
	cached := filepath.Join(cache, filename)
	if bundleRuntime {
		if err = plainParents(cache); err != nil {
			return err
		}
		if err = os.MkdirAll(cache, 0755); err != nil {
			return err
		}
		if err = download(cached, url, pin.SHA256, offline); err != nil {
			return err
		}
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
	if bundleRuntime {
		if err = runtimebundle.ExtractJar(cached, pin.Archive, root); err != nil {
			return err
		}
		if target == "windows-amd64" {
			runtimeFile := filepath.Join(cache, "Microsoft.VCLibs.x64.14.00.Desktop.appx")
			if err = download(runtimeFile, pins.WindowsVCRuntime.URL, pins.WindowsVCRuntime.SHA256, offline); err != nil {
				return err
			}
			if err = runtimebundle.ExtractVCRuntime(runtimeFile, pins.WindowsVCRuntime, root); err != nil {
				return err
			}
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
	for _, doc := range []string{"README.md", "README.ko.md", "docs/cli.md", "docs/ko/cli.md", "docs/CLI-VERIFICATION.md", "docs/ko/CLI-VERIFICATION.md", "docs/PROVENANCE.md", "docs/ko/PROVENANCE.md", "docs/ai-providers.md", "docs/ko/ai-providers.md"} {
		if err = copyFile(root, doc, filepath.Join(repo, filepath.FromSlash(doc)), 0644); err != nil {
			return err
		}
	}
	if err = copyTree(root, "docs/evidence/standalone-cli", filepath.Join(repo, "docs", "evidence", "standalone-cli")); err != nil {
		return err
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
	mode := "download"
	if bundleRuntime {
		mode = "bundled"
	}
	m := manifest{FormatVersion: 2, Target: target, SourceRevision: strings.TrimSpace(string(revision)), SourceDirty: len(status) > 0, PostgresVersion: pins.PostgresVersion, PostgresURL: url, PostgresSHA256: pin.SHA256, Publisher: pins.Publisher, RuntimeMode: mode, RuntimeOnly: runtimeOnly, Files: map[string]string{}}
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

func fileHash(name string) (string, error) { return runtimebundle.FileHash(name) }
func download(name, url, sum string, offline bool) error {
	return runtimebundle.Download(context.Background(), name, url, sum, offline)
}
func safeName(name string) (string, error)          { return runtimebundle.SafeName(name) }
func extractTar(t *tar.Reader, root *os.Root) error { return runtimebundle.ExtractTar(t, root) }
func writeFile(root *os.Root, name string, data []byte, mode fs.FileMode) error {
	return runtimebundle.WriteFile(root, name, data, mode)
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
