//go:build windows

package portable

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPostgresWindowsUnicodeCommand(t *testing.T) {
	home := filepath.Join(t.TempDir(), "portable state 한글")
	bin := filepath.Join(home, "runtime", "postgres", "bin")
	data := filepath.Join(home, "postgres")
	private := filepath.Join(home, "private")
	for _, dir := range []string{bin, data, private} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	exe := filepath.Join(bin, "pg_ctl.exe")
	password := filepath.Join(private, "initdb-password")
	for _, path := range []string{exe, password} {
		if err := os.WriteFile(path, []byte("fixture"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	log := filepath.Join(private, "postgres.log") // PG creates this file later.
	cmd, err := newPostgresCommand(context.Background(), exe, "-D", data, "-l", log, "--pwfile="+password, "-o", "-h 127.0.0.1", "start")
	if err != nil {
		t.Fatal(err)
	}
	paths := []struct{ actual, expected string }{
		{cmd.Path, exe}, {cmd.Dir, bin}, {cmd.Args[2], data},
		{filepath.Dir(cmd.Args[4]), private}, {strings.TrimPrefix(cmd.Args[5], "--pwfile="), password},
	}
	for _, pair := range paths {
		for _, c := range pair.actual {
			if c > 127 {
				t.Fatalf("non-ASCII PostgreSQL path: %q", pair.actual)
			}
		}
		actual, err := os.Stat(pair.actual)
		if err != nil {
			t.Fatal(err)
		}
		expected, err := os.Stat(pair.expected)
		if err != nil {
			t.Fatal(err)
		}
		if !os.SameFile(actual, expected) {
			t.Fatalf("path alias changed identity: %q != %q", pair.actual, pair.expected)
		}
	}
	if cmd.Args[6] != "-o" || cmd.Args[7] != "-h 127.0.0.1" || cmd.Args[8] != "start" {
		t.Fatalf("non-path options changed: %q", cmd.Args)
	}
}

func TestPostgresWindowsASCIIWithoutShortNames(t *testing.T) {
	path := `C:\portable long folder\missing-data\postgres.log`
	got, err := postgresPath(path)
	if err != nil || got != path {
		t.Fatalf("ASCII paths must not require short names: %q, %v", got, err)
	}
}

func TestPostgresWindowsRejectsUnicodeWithoutShortNames(t *testing.T) {
	if err := requirePostgresASCIIPath(`C:\portable state 한글\postgres`); err == nil || !strings.Contains(err.Error(), "8.3") || !strings.Contains(err.Error(), "ASCII") {
		t.Fatalf("missing actionable unsupported-path error: %v", err)
	}
}
