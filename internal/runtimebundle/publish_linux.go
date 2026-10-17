//go:build linux

package runtimebundle

import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
)

func publishNew(from, to string) error {
	return atomicPublicationError(unix.Renameat2(unix.AT_FDCWD, from, unix.AT_FDCWD, to, unix.RENAME_NOREPLACE))
}
func atomicPublicationError(err error) error {
	if errors.Is(err, unix.EINVAL) || errors.Is(err, unix.ENOTSUP) || errors.Is(err, unix.ENOSYS) {
		return fmt.Errorf("filesystem does not support atomic portable runtime publication; use a native Linux filesystem instead of a Windows-mounted WSL folder: %w", err)
	}
	return err
}
