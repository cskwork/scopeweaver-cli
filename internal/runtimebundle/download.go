// Package runtimebundle verifies and extracts the pinned portable PostgreSQL runtime.
package runtimebundle

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//go:embed postgres-sources.json
var sourceBytes []byte

// Artifact identifies a checksum-pinned publisher archive.
type Artifact struct {
	Artifact string `json:"artifact"`
	Archive  string `json:"archive"`
	SHA256   string `json:"sha256"`
}

// VCRuntime identifies the app-local DLLs needed by Windows PostgreSQL.
type VCRuntime struct {
	Version        string   `json:"version"`
	URL            string   `json:"url"`
	SHA256         string   `json:"sha256"`
	ChecksumSource string   `json:"checksum_source"`
	Files          []string `json:"files"`
}

// Sources contains the pins embedded in this CLI build.
type Sources struct {
	PostgresVersion  string              `json:"postgres_version"`
	PublisherVersion string              `json:"publisher_version"`
	Publisher        string              `json:"publisher"`
	Repository       string              `json:"repository"`
	Targets          map[string]Artifact `json:"targets"`
	WindowsVCRuntime VCRuntime           `json:"windows_vc_runtime"`
}

func SourceBytes() []byte { return append([]byte(nil), sourceBytes...) }
func PinnedSources() (Sources, error) {
	var p Sources
	err := json.Unmarshal(sourceBytes, &p)
	return p, err
}
func (p Sources) PostgresArtifact(target string) (filename, source string, err error) {
	pin, ok := p.Targets[target]
	if !ok {
		return "", "", fmt.Errorf("unsupported portable runtime target %q", target)
	}
	filename = "embedded-postgres-binaries-" + pin.Artifact + "-" + p.PublisherVersion + ".jar"
	source = p.Repository + "embedded-postgres-binaries-" + pin.Artifact + "/" + p.PublisherVersion + "/" + filename
	return filename, source, nil
}
func FileHash(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func verify(name, want string) error {
	info, err := os.Lstat(name)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("runtime archive is not a regular file: %s", name)
	}
	got, err := FileHash(name)
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("SHA256 mismatch for %s: got %s, want %s", name, got, want)
	}
	return nil
}
func secureURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" {
		return fmt.Errorf("runtime download requires an HTTPS source: %q", raw)
	}
	return nil
}

// Download verifies cached bytes or atomically publishes a fully verified download.
// Partial download files are always removed; neither cache nor installed runtime is overwritten.
func Download(ctx context.Context, name, source, sum string, offline bool) error {
	return downloadWithClient(ctx, &http.Client{Timeout: 5 * time.Minute, CheckRedirect: func(r *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("runtime download exceeded redirect limit")
		}
		return secureURL(r.URL.String())
	}}, name, source, sum, offline)
}
func downloadWithClient(ctx context.Context, client *http.Client, name, source, sum string, offline bool) error {
	if _, err := os.Lstat(name); err == nil {
		return verify(name, sum)
	} else if !os.IsNotExist(err) {
		return err
	}
	if offline {
		return fmt.Errorf("offline runtime cache missing: %s", name)
	}
	if err := secureURL(source); err != nil {
		return err
	}
	if decoded, err := hex.DecodeString(sum); err != nil || len(decoded) != sha256.Size {
		return errors.New("invalid pinned runtime SHA256")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download portable runtime from %s (check network access): %w", source, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download portable runtime from %s: HTTP %s", source, response.Status)
	}
	f, err := os.CreateTemp(filepath.Dir(name), ".download-*")
	if err != nil {
		return fmt.Errorf("write portable runtime next to CLI (check folder permissions): %w", err)
	}
	temporary := f.Name()
	defer os.Remove(temporary)
	size, copyErr := io.Copy(f, io.LimitReader(response.Body, (512<<20)+1))
	closeErr := f.Close()
	if copyErr != nil {
		return fmt.Errorf("download portable runtime interrupted: %w", copyErr)
	}
	if closeErr != nil {
		return closeErr
	}
	if size > 512<<20 {
		return errors.New("runtime download exceeds 512 MiB limit")
	}
	if err = verify(temporary, sum); err != nil {
		return err
	}
	if _, err = os.Lstat(name); err == nil {
		return errors.New("runtime cache appeared while downloading")
	} else if !os.IsNotExist(err) {
		return err
	}
	return publishNew(temporary, name)
}
func CheckParents(name string) error {
	for p := filepath.Clean(name); ; p = filepath.Dir(p) {
		info, err := os.Lstat(p)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink runtime parent is not allowed: %s", p)
		}
		if err == nil && !info.IsDir() {
			return fmt.Errorf("portable runtime path is not a directory: %s (check folder permissions and choose a writable portable folder)", p)
		}
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("inspect portable runtime folder (check folder permissions): %w", err)
		}
		if filepath.Dir(p) == p {
			return nil
		}
	}
}

// Validate checks for a complete existing runtime without executing or downloading it.
func Validate(executableRoot, target string) error {
	if err := CheckParents(filepath.Join(executableRoot, "runtime", "postgres", "bin")); err != nil {
		return err
	}
	suffix := ""
	if strings.HasPrefix(target, "windows-") {
		suffix = ".exe"
	}
	for _, name := range []string{"initdb", "pg_ctl", "postgres"} {
		full := filepath.Join(executableRoot, "runtime", "postgres", "bin", name+suffix)
		info, err := os.Lstat(full)
		if err != nil {
			return fmt.Errorf("portable PostgreSQL is missing %s: %w", full, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("portable PostgreSQL is not a regular file: %s", full)
		}
		if suffix == "" && runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
			return fmt.Errorf("portable PostgreSQL is not executable: %s", full)
		}
	}
	return nil
}

// Ensure installs only when runtime/postgres does not exist. Callers must hold
// the executable-root install lock across this call. Existing folders are preserved.
func Ensure(ctx context.Context, executableRoot, target string) error {
	pins, err := PinnedSources()
	if err != nil {
		return err
	}
	return ensureWithSources(ctx, executableRoot, target, pins, Download)
}

type downloader func(context.Context, string, string, string, bool) error

func ensureWithSources(ctx context.Context, executableRoot, target string, pins Sources, download downloader) error {
	installed := filepath.Join(executableRoot, "runtime", "postgres")
	if err := CheckParents(installed); err != nil {
		return err
	}
	if _, err := os.Lstat(installed); err == nil {
		if err = Validate(executableRoot, target); err != nil {
			return fmt.Errorf("existing portable runtime is incomplete; preserve this folder and restore a complete matching runtime: %w", err)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	filename, source, err := pins.PostgresArtifact(target)
	if err != nil {
		return err
	}
	runtimeDir := filepath.Dir(installed)
	if err = os.MkdirAll(runtimeDir, 0755); err != nil {
		return fmt.Errorf("create portable runtime next to CLI (check folder permissions): %w", err)
	}
	staging, err := os.MkdirTemp(runtimeDir, ".install-*")
	if err != nil {
		return fmt.Errorf("stage portable runtime next to CLI (check folder permissions): %w", err)
	}
	defer os.RemoveAll(staging) // Only this call's scratch directory is removed.
	root, err := os.OpenRoot(staging)
	if err != nil {
		return err
	}
	defer root.Close()
	cached := filepath.Join(staging, filename)
	pin := pins.Targets[target]
	if err = download(ctx, cached, source, pin.SHA256, false); err != nil {
		return err
	}
	// Reverify at the extraction boundary, including custom downloaders in tests.
	if err = verify(cached, pin.SHA256); err != nil {
		return err
	}
	if err = ExtractJar(cached, pin.Archive, root); err != nil {
		return fmt.Errorf("extract portable PostgreSQL: %w", err)
	}
	if target == "windows-amd64" {
		vc := filepath.Join(staging, "vc-runtime.appx")
		if err = download(ctx, vc, pins.WindowsVCRuntime.URL, pins.WindowsVCRuntime.SHA256, false); err != nil {
			return err
		}
		if err = verify(vc, pins.WindowsVCRuntime.SHA256); err != nil {
			return err
		}
		if err = ExtractVCRuntime(vc, pins.WindowsVCRuntime, root); err != nil {
			return err
		}
	}
	if err = Validate(staging, target); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	// Same-filesystem rename exposes PostgreSQL and its app-local DLLs together.
	if err = publishNew(filepath.Join(staging, "runtime", "postgres"), installed); err != nil {
		return fmt.Errorf("publish portable runtime without replacing existing folders: %w", err)
	}
	return nil
}
