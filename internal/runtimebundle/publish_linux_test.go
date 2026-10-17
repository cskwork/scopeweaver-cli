//go:build linux

package runtimebundle

import (
	"errors"
	"golang.org/x/sys/unix"
	"strings"
	"testing"
)

func TestUnsupportedAtomicPublicationProvidesFilesystemGuidance(t *testing.T) {
	for _, cause := range []error{unix.EINVAL, unix.ENOTSUP, unix.ENOSYS} {
		got := atomicPublicationError(cause)
		if !errors.Is(got, cause) || !strings.Contains(got.Error(), "use a native Linux filesystem") {
			t.Fatalf("unsupported filesystem error lost cause or guidance: %v", got)
		}
	}
	// Destination conflicts and permission errors retain their distinct meaning.
	for _, cause := range []error{unix.EEXIST, unix.EACCES, nil} {
		if got := atomicPublicationError(cause); got != cause {
			t.Fatalf("changed unrelated publication error %v into %v", cause, got)
		}
	}
}
