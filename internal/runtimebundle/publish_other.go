//go:build !windows && !linux && !darwin

package runtimebundle

import "errors"

func publishNew(from, to string) error {
	return errors.New("portable runtime installation is unsupported on this platform")
}
