package portable

import (
	"context"
	"os"
	"os/exec"
)

func newPostgresCommand(ctx context.Context, exe string, args ...string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, exe, args...)
	if err := preparePostgresCommand(cmd); err != nil {
		return nil, err
	}
	cmd.Env = isolatedEnv(os.Environ())
	hideCommand(cmd)
	return cmd, nil
}
