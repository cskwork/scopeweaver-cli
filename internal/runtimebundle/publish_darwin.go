//go:build darwin

package runtimebundle

import "golang.org/x/sys/unix"

func publishNew(from, to string) error { return unix.RenamexNp(from, to, unix.RENAME_EXCL) }
