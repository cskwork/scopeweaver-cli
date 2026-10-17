//go:build !windows

package portable

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func lockFile(f *os.File) error { return syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) }
func lockBusy(err error) bool {
	return errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN)
}
func detach(cmd *exec.Cmd)         { cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} }
func hideCommand(cmd *exec.Cmd)    {}
func protectDir(path string) error { return os.Chmod(path, 0700) }

func publishSeed(temporary, target string) error {
	// A same-directory hard link atomically publishes complete content and fails
	// if the target already exists; the caller removes the temporary afterward.
	return os.Link(temporary, target)
}
