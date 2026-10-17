package portable

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRuntimeInstallLockSharedAcrossHomes(t *testing.T) {
	root := t.TempDir()
	target := runtime.GOOS + "-" + runtime.GOARCH
	var calls atomic.Int32
	var active atomic.Int32
	var overlap atomic.Bool
	install := func(ctx context.Context, root, target string) error {
		if active.Add(1) != 1 {
			overlap.Store(true)
		}
		defer active.Add(-1)
		bin := filepath.Join(root, "runtime", "postgres", "bin")
		if _, err := os.Stat(bin); err == nil {
			return nil
		}
		calls.Add(1)
		time.Sleep(30 * time.Millisecond)
		if err := os.MkdirAll(bin, 0755); err != nil {
			return err
		}
		suffix := ""
		if runtime.GOOS == "windows" {
			suffix = ".exe"
		}
		for _, name := range []string{"initdb", "pg_ctl", "postgres"} {
			if err := os.WriteFile(filepath.Join(bin, name+suffix), []byte("fixture"), 0755); err != nil {
				return err
			}
		}
		return nil
	}
	var group sync.WaitGroup
	errs := make(chan error, 6)
	for i := 0; i < 6; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			home := filepath.Join(root, fmt.Sprint("home-", i))
			if err := prepareHome(home); err != nil {
				errs <- err
				return
			}
			errs <- ensurePostgresWith(context.Background(), root, target, install)
		}(i)
	}
	group.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if overlap.Load() || calls.Load() != 1 {
		t.Fatalf("installs=%d overlapping=%v", calls.Load(), overlap.Load())
	}
}
func TestRuntimeInstallLockHonorsCancellation(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "runtime"), 0755)
	lock, err := acquireLock(context.Background(), filepath.Join(root, "runtime", "install.lock"))
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	err = ensurePostgresWith(ctx, root, runtime.GOOS+"-"+runtime.GOARCH, func(context.Context, string, string) error { t.Fatal("acquired held install lock"); return nil })
	if err == nil {
		t.Fatal("lock wait ignored cancellation")
	}
}

func TestRuntimeInstallRejectsCompleteSymlinkBeforeFastPath(t *testing.T) {
	root := t.TempDir()
	real := t.TempDir()
	bin := filepath.Join(real, "postgres", "bin")
	os.MkdirAll(bin, 0755)
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	for _, name := range []string{"initdb", "pg_ctl", "postgres"} {
		if err := os.WriteFile(filepath.Join(bin, name+suffix), []byte("preserve"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(real, filepath.Join(root, "runtime")); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("Windows symlink privilege unavailable: %v", err)
		}
		t.Fatal(err)
	}
	err := ensurePostgresWith(context.Background(), root, runtime.GOOS+"-"+runtime.GOARCH, func(context.Context, string, string) error {
		t.Fatal("installer reached redirected runtime")
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "symlink runtime parent") {
		t.Fatalf("error=%v", err)
	}
	if _, err := os.Stat(filepath.Join(real, "install.lock")); !os.IsNotExist(err) {
		t.Fatal("lock written through symlink")
	}
}
