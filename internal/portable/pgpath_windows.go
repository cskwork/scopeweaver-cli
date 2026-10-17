//go:build windows

package portable

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// PostgreSQL's Windows tools use narrow paths, including getcwd internally.
// Give every tool an ASCII alias and cwd without changing the actual storage.
func preparePostgresCommand(cmd *exec.Cmd) error {
	exe, err := postgresPath(cmd.Path)
	if err != nil {
		return err
	}
	cmd.Path = exe
	cmd.Args = append([]string(nil), cmd.Args...)
	cmd.Args[0] = exe
	cmd.Dir = filepath.Dir(exe)
	for i := 1; i < len(cmd.Args); i++ {
		switch {
		case cmd.Args[i] == "-D" || cmd.Args[i] == "-l":
			if i+1 >= len(cmd.Args) {
				return fmt.Errorf("missing PostgreSQL path after %s", cmd.Args[i])
			}
			i++
			cmd.Args[i], err = postgresPath(cmd.Args[i])
		case strings.HasPrefix(cmd.Args[i], "--pwfile="):
			var path string
			path, err = postgresPath(strings.TrimPrefix(cmd.Args[i], "--pwfile="))
			cmd.Args[i] = "--pwfile=" + path
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func requirePostgresASCIIPath(path string) error {
	for _, c := range path {
		if c > 127 {
			return fmt.Errorf("bundled PostgreSQL cannot use Unicode path %q: this volume has no usable ASCII 8.3 alias; use an ASCII portable folder and home (no system settings were changed)", path)
		}
	}
	return nil
}

func postgresPath(path string) (string, error) {
	if requirePostgresASCIIPath(path) == nil {
		return path, nil
	}
	// Log files need not exist yet; their existing private parent has an alias.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		parent, err := postgresPath(filepath.Dir(path))
		if err != nil {
			return "", err
		}
		alias := filepath.Join(parent, filepath.Base(path))
		if err := requirePostgresASCIIPath(alias); err != nil {
			return "", err
		}
		return alias, nil
	} else if err != nil {
		return "", err
	}
	input, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	size, err := windows.GetShortPathName(input, nil, 0)
	if err != nil {
		return "", fmt.Errorf("resolve PostgreSQL ASCII path: %w", err)
	}
	for {
		output := make([]uint16, size)
		n, err := windows.GetShortPathName(input, &output[0], uint32(len(output)))
		if err != nil {
			return "", fmt.Errorf("resolve PostgreSQL ASCII path: %w", err)
		}
		if n >= uint32(len(output)) {
			size = n
			continue
		}
		alias := windows.UTF16ToString(output[:n])
		if err := requirePostgresASCIIPath(alias); err != nil {
			return "", err
		}
		original, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		resolved, err := os.Stat(alias)
		if err != nil {
			return "", err
		}
		if !os.SameFile(original, resolved) {
			return "", fmt.Errorf("PostgreSQL path alias does not match its original: %q", path)
		}
		return alias, nil
	}
}
