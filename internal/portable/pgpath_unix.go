//go:build !windows

package portable

import "os/exec"

func preparePostgresCommand(cmd *exec.Cmd) error { return nil }
