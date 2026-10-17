package runtimebundle

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/ulikunitz/xz"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func digest(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type interruptedReader struct{ read bool }

func (r *interruptedReader) Read(p []byte) (int, error) {
	if r.read {
		return 0, io.ErrUnexpectedEOF
	}
	r.read = true
	return copy(p, "partial"), nil
}
func (r *interruptedReader) Close() error { return nil }
func TestDownloadFailureRetryAndHashBoundary(t *testing.T) {
	for _, kind := range []string{"interrupted", "hash", "network", "http"} {
		t.Run(kind, func(t *testing.T) {
			dest := filepath.Join(t.TempDir(), "runtime.jar")
			attempt := 0
			client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				attempt++
				if attempt == 1 && kind == "network" {
					return nil, errors.New("offline fixture")
				}
				body := io.NopCloser(strings.NewReader("verified"))
				status := 200
				if attempt == 1 {
					switch kind {
					case "interrupted":
						body = &interruptedReader{}
					case "hash":
						body = io.NopCloser(strings.NewReader("corrupt"))
					case "http":
						status = 503
					}
				}
				return &http.Response{StatusCode: status, Status: fmt.Sprint(status), Body: body, Header: make(http.Header), Request: r}, nil
			})}
			if err := downloadWithClient(context.Background(), client, dest, "https://fixture.invalid/runtime", digest([]byte("verified")), false); err == nil {
				t.Fatal("failure accepted")
			}
			if _, err := os.Stat(dest); !os.IsNotExist(err) {
				t.Fatalf("failed bytes published: %v", err)
			}
			leftovers, _ := filepath.Glob(filepath.Join(filepath.Dir(dest), ".download-*"))
			if len(leftovers) != 0 {
				t.Fatalf("partial download retained: %v", leftovers)
			}
			if err := downloadWithClient(context.Background(), client, dest, "https://fixture.invalid/runtime", digest([]byte("verified")), false); err != nil {
				t.Fatal(err)
			}
			if got, _ := os.ReadFile(dest); string(got) != "verified" {
				t.Fatalf("retry=%q", got)
			}
		})
	}
}
func TestDownloadRequiresHTTPSAndValidPinWithoutNetwork(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) { calls++; return nil, errors.New("unexpected network") })}
	for _, source := range []string{"http://fixture.invalid/runtime", "file:///runtime", "https://user:password@fixture.invalid/runtime", "https://fixture.invalid/runtime#fragment"} {
		if err := downloadWithClient(context.Background(), client, filepath.Join(t.TempDir(), "artifact"), source, digest(nil), false); err == nil {
			t.Fatalf("accepted %s", source)
		}
	}
	if err := downloadWithClient(context.Background(), client, filepath.Join(t.TempDir(), "artifact"), "https://fixture.invalid/runtime", "bad", false); err == nil {
		t.Fatal("accepted invalid pin")
	}
	if calls != 0 {
		t.Fatalf("unexpected network calls=%d", calls)
	}
}
func TestPinnedSourcesUseHTTPS(t *testing.T) {
	pins, err := PinnedSources()
	if err != nil {
		t.Fatal(err)
	}
	for target, pin := range pins.Targets {
		_, source, err := pins.PostgresArtifact(target)
		if err != nil {
			t.Fatal(err)
		}
		if err = secureURL(source); err != nil {
			t.Fatal(err)
		}
		if len(pin.SHA256) != 64 {
			t.Fatal("invalid pin")
		}
	}
	if err = secureURL(pins.WindowsVCRuntime.URL); err != nil {
		t.Fatal(err)
	}
}
func fixtureArchive(t *testing.T, unsafe bool) []byte {
	t.Helper()
	var txz bytes.Buffer
	x, err := xz.NewWriter(&txz)
	if err != nil {
		t.Fatal(err)
	}
	tarWriter := tar.NewWriter(x)
	names := []string{"bin/initdb", "bin/pg_ctl", "bin/postgres"}
	if unsafe {
		names = append(names, "../../outside")
	}
	for _, name := range names {
		if err = tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0755, Size: 1, Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err = tarWriter.Write([]byte("x")); err != nil {
			t.Fatal(err)
		}
	}
	if err = tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err = x.Close(); err != nil {
		t.Fatal(err)
	}
	var jar bytes.Buffer
	z := zip.NewWriter(&jar)
	member, err := z.Create("postgres.txz")
	if err != nil {
		t.Fatal(err)
	}
	member.Write(txz.Bytes())
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	return jar.Bytes()
}
func fixturePins(data []byte) Sources {
	return Sources{PublisherVersion: "fixture", Repository: "https://fixture.invalid/", Targets: map[string]Artifact{"linux-amd64": {Artifact: "fixture", Archive: "postgres.txz", SHA256: digest(data)}}}
}
func fixtureDownload(data []byte) downloader {
	return func(_ context.Context, dest, _, _ string, _ bool) error { return os.WriteFile(dest, data, 0600) }
}
func TestInstallPublishesCompleteRuntimeAndReusesWithoutNetwork(t *testing.T) {
	data := fixtureArchive(t, false)
	root := t.TempDir()
	pins := fixturePins(data)
	if err := ensureWithSources(context.Background(), root, "linux-amd64", pins, fixtureDownload(data)); err != nil {
		t.Fatal(err)
	}
	if err := Validate(root, "linux-amd64"); err != nil {
		t.Fatal(err)
	}
	if err := ensureWithSources(context.Background(), root, "linux-amd64", pins, func(context.Context, string, string, string, bool) error {
		t.Fatal("existing runtime downloaded again")
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	leftovers, _ := filepath.Glob(filepath.Join(root, "runtime", ".install-*"))
	if len(leftovers) != 0 {
		t.Fatalf("staging retained: %v", leftovers)
	}
}
func TestInstallFailuresPreserveDataAndAllowRetry(t *testing.T) {
	good := fixtureArchive(t, false)
	for _, kind := range []string{"hash", "archive", "interrupted", "cancelled"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			sentinel := filepath.Join(root, "user-data")
			os.WriteFile(sentinel, []byte("preserve"), 0600)
			payload := good
			pins := fixturePins(good)
			download := fixtureDownload(payload)
			ctx := context.Background()
			switch kind {
			case "hash":
				download = fixtureDownload([]byte("corrupt"))
			case "archive":
				payload = fixtureArchive(t, true)
				pins = fixturePins(payload)
				download = fixtureDownload(payload)
			case "interrupted":
				download = func(context.Context, string, string, string, bool) error { return io.ErrUnexpectedEOF }
			case "cancelled":
				cancelled, cancel := context.WithCancel(ctx)
				cancel()
				ctx = cancelled
			}
			if err := ensureWithSources(ctx, root, "linux-amd64", pins, download); err == nil {
				t.Fatal("failed install succeeded")
			}
			if _, err := os.Stat(filepath.Join(root, "runtime", "postgres")); !os.IsNotExist(err) {
				t.Fatalf("partial runtime exposed: %v", err)
			}
			if got, _ := os.ReadFile(sentinel); string(got) != "preserve" {
				t.Fatal("user data changed")
			}
			if err := ensureWithSources(context.Background(), root, "linux-amd64", fixturePins(good), fixtureDownload(good)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestExistingIncompleteRuntimeIsPreserved(t *testing.T) {
	root := t.TempDir()
	existing := filepath.Join(root, "runtime", "postgres")
	os.MkdirAll(existing, 0755)
	sentinel := filepath.Join(existing, "existing")
	os.WriteFile(sentinel, []byte("keep"), 0600)
	err := ensureWithSources(context.Background(), root, "linux-amd64", fixturePins(nil), func(context.Context, string, string, string, bool) error {
		t.Fatal("download into existing runtime")
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "preserve this folder") {
		t.Fatalf("error=%v", err)
	}
	if got, _ := os.ReadFile(sentinel); string(got) != "keep" {
		t.Fatal("existing runtime changed")
	}
}
func TestRuntimeFolderWriteFailureIsActionable(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "runtime"), []byte("keep"), 0600)
	err := Ensure(context.Background(), root, "linux-amd64")
	if err == nil || !strings.Contains(err.Error(), "folder permissions") {
		t.Fatalf("error=%v", err)
	}
}

func TestAtomicPublishNeverReplacesExistingFileOrDirectory(t *testing.T) {
	for _, directory := range []bool{false, true} {
		t.Run(fmt.Sprint("directory=", directory), func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(root, "source")
			target := filepath.Join(root, "target")
			if directory {
				os.Mkdir(source, 0755)
				os.Mkdir(target, 0755)
			} else {
				os.WriteFile(source, []byte("source"), 0600)
				os.WriteFile(target, []byte("target"), 0600)
			}
			if err := publishNew(source, target); err == nil {
				t.Fatal("replaced existing destination")
			}
			if _, err := os.Stat(source); err != nil {
				t.Fatal("source changed after failed publish")
			}
			if !directory {
				if got, _ := os.ReadFile(target); string(got) != "target" {
					t.Fatal("destination changed")
				}
			}
		})
	}
}

func TestValidateRejectsSymlinkParentsWithCompleteBinaries(t *testing.T) {
	for _, level := range []string{"runtime", "postgres", "bin"} {
		t.Run(level, func(t *testing.T) {
			root := t.TempDir()
			real := t.TempDir()
			bin := real
			link := filepath.Join(root, "runtime")
			switch level {
			case "runtime":
				bin = filepath.Join(real, "postgres", "bin")
			case "postgres":
				link = filepath.Join(root, "runtime", "postgres")
				bin = filepath.Join(real, "bin")
			case "bin":
				link = filepath.Join(root, "runtime", "postgres", "bin")
			}
			if err := os.MkdirAll(filepath.Dir(link), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(bin, 0755); err != nil {
				t.Fatal(err)
			}
			suffix := ""
			if runtime.GOOS == "windows" {
				suffix = ".exe"
			}
			for _, name := range []string{"initdb", "pg_ctl", "postgres"} {
				if err := os.WriteFile(filepath.Join(bin, name+suffix), []byte("preserve"), 0755); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Symlink(real, link); err != nil {
				if runtime.GOOS == "windows" {
					t.Skipf("Windows symlink privilege unavailable: %v", err)
				}
				t.Fatal(err)
			}
			if err := Validate(root, runtime.GOOS+"-"+runtime.GOARCH); err == nil || !strings.Contains(err.Error(), "symlink runtime parent") {
				t.Fatalf("complete runtime through symlink accepted: %v", err)
			}
			if got, _ := os.ReadFile(filepath.Join(bin, "postgres"+suffix)); string(got) != "preserve" {
				t.Fatal("linked runtime changed")
			}
		})
	}
}
