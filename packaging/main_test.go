package main

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSafeName(t *testing.T) {
	for _, name := range []string{"../escape", "/absolute", "C:/escape", "C:escape", "bin\\escape", "bin/../../../escape", ".", "..", "NUL"} {
		if _, err := safeName(name); err == nil {
			t.Errorf("accepted %q", name)
		}
	}
	for _, name := range []string{"bin/postgres.exe", "lib/libpq.so.5", "share/postgresql/postgres.bki"} {
		if _, err := safeName(name); err != nil {
			t.Errorf("rejected %q: %v", name, err)
		}
	}
}

func TestExtractTarRejectsEscapesAndDuplicate(t *testing.T) {
	cases := []struct {
		name    string
		headers []*tar.Header
		bad     bool
	}{
		{"regular", []*tar.Header{{Name: "bin/postgres", Typeflag: tar.TypeReg, Mode: 0755, Size: 1}}, false},
		{"traversal", []*tar.Header{{Name: "../../outside", Typeflag: tar.TypeReg, Mode: 0644}}, true},
		{"windows-path", []*tar.Header{{Name: "bin\\..\\outside", Typeflag: tar.TypeReg, Mode: 0644}}, true},
		{"symlink-escape", []*tar.Header{{Name: "lib/outside", Typeflag: tar.TypeSymlink, Linkname: "../../../outside"}}, true},
		{"absolute-symlink", []*tar.Header{{Name: "lib/outside", Typeflag: tar.TypeSymlink, Linkname: "/tmp/outside"}}, true},
		{"hardlink-escape", []*tar.Header{{Name: "lib/outside", Typeflag: tar.TypeLink, Linkname: "../../outside"}}, true},
		{"device", []*tar.Header{{Name: "device", Typeflag: tar.TypeChar}}, true},
		{"duplicate", []*tar.Header{{Name: "bin/postgres", Typeflag: tar.TypeReg}, {Name: "bin/postgres", Typeflag: tar.TypeReg}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var data bytes.Buffer
			w := tar.NewWriter(&data)
			for _, h := range tc.headers {
				if err := w.WriteHeader(h); err != nil {
					t.Fatal(err)
				}
				if h.Size > 0 {
					w.Write(bytes.Repeat([]byte{'x'}, int(h.Size)))
				}
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			dest := t.TempDir()
			root, err := os.OpenRoot(dest)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			err = extractTar(tar.NewReader(&data), root)
			if (err != nil) != tc.bad {
				t.Fatalf("error=%v, want rejection=%v", err, tc.bad)
			}
			if !tc.bad {
				got, err := os.ReadFile(filepath.Join(dest, "runtime/postgres/bin/postgres"))
				if err != nil || string(got) != "x" {
					t.Fatalf("extracted content=%q, error=%v", got, err)
				}
			}
		})
	}
}

func TestDownloadRejectsCorruptCacheOffline(t *testing.T) {
	file := filepath.Join(t.TempDir(), "postgres.jar")
	if err := os.WriteFile(file, []byte("corrupt"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := download(file, "https://invalid.example/unused", "0000000000000000000000000000000000000000000000000000000000000000", true); err == nil {
		t.Fatal("accepted corrupt cache")
	}
	got, _ := os.ReadFile(file)
	if string(got) != "corrupt" {
		t.Fatal("modified corrupt cache")
	}
}

func TestExtractTarMaterializesInternalLibraryLinks(t *testing.T) {
	var data bytes.Buffer
	w := tar.NewWriter(&data)
	for _, h := range []*tar.Header{
		{Name: "lib/libexample.so", Typeflag: tar.TypeSymlink, Linkname: "libexample.so.1"},
		{Name: "lib/libexample.so.1", Typeflag: tar.TypeReg, Mode: 0755, Size: 1},
		{Name: "lib/hardlink", Typeflag: tar.TypeLink, Linkname: "lib/libexample.so.1"},
	} {
		if err := w.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if h.Size > 0 {
			w.Write([]byte("x"))
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	root, err := os.OpenRoot(dest)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err = extractTar(tar.NewReader(&data), root); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"libexample.so", "libexample.so.1", "hardlink"} {
		p := filepath.Join(dest, "runtime/postgres/lib", name)
		info, err := os.Lstat(p)
		if err != nil || !info.Mode().IsRegular() {
			t.Fatalf("%s is not materialized: %v", name, err)
		}
		got, err := os.ReadFile(p)
		if err != nil || string(got) != "x" {
			t.Fatalf("%s has content %q, error=%v", name, got, err)
		}
	}
}

func TestPreservedLicenseHashes(t *testing.T) {
	var sources struct {
		Sources []struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
		} `json:"license_sources"`
	}
	raw, err := assets.ReadFile("licenses/SOURCES.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &sources); err != nil {
		t.Fatal(err)
	}
	if len(sources.Sources) == 0 {
		t.Fatal("missing license provenance")
	}
	for _, source := range sources.Sources {
		data, err := assets.ReadFile("licenses/" + source.Path)
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != source.SHA256 {
			t.Errorf("license hash mismatch: %s", source.Path)
		}
	}
}
