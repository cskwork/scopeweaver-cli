package portable

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// seedSkills initializes each independent home from the shipped bundle once.
// Completed homes retain edits and deliberate deletions across restart/update;
// interrupted attempts never overwrite a preexisting file.
func seedSkills(ctx context.Context, bundle, home string) (string, error) {
	root := filepath.Join(home, "skills")
	marker := filepath.Join(privateDir(home), "skills-seeded.json")
	if b, err := os.ReadFile(marker); err == nil {
		var state struct {
			Complete bool `json:"complete"`
		}
		if json.Unmarshal(b, &state) != nil || !state.Complete {
			return "", errors.New("invalid private skills initialization marker; restore it from this home's backup")
		}
		if err = ensurePlainDirectory(root); err != nil {
			return "", err
		}
		return root, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	source, err := os.Stat(bundle)
	if err != nil || !source.IsDir() {
		return "", fmt.Errorf("bundled skills are missing at %s; unpack the complete portable release", bundle)
	}
	if err = ensurePlainDirectory(root); err != nil {
		return "", err
	}
	err = filepath.WalkDir(bundle, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		rel, err := filepath.Rel(bundle, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("bundled skill symlink is not portable: %s", rel)
		}
		target := filepath.Join(root, rel)
		if entry.IsDir() {
			return ensurePlainDirectory(target)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("bundled skill is not a regular file: %s", rel)
		}
		if existing, err := os.Lstat(target); err == nil {
			if !existing.Mode().IsRegular() {
				return fmt.Errorf("existing skill path is not a regular file: %s", rel)
			}
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		defer input.Close()
		output, err := os.CreateTemp(filepath.Dir(target), ".scopeweaver-seed-*")
		if err != nil {
			return err
		}
		temporary := output.Name()
		defer os.Remove(temporary)
		_, copyErr := io.Copy(output, input)
		modeErr := output.Chmod(info.Mode().Perm())
		syncErr := output.Sync()
		closeErr := output.Close()
		if err = errors.Join(copyErr, modeErr, syncErr, closeErr); err != nil {
			return err
		}
		// Publish complete content atomically without replacing a user file that
		// appeared since the existence check. A crash leaves only a temporary.
		if err = publishSeed(temporary, target); os.IsExist(err) {
			return nil
		}
		return err
	})
	if err != nil {
		return "", fmt.Errorf("initialize home skills: %w", err)
	}
	if err = writeJSONFile(marker, map[string]bool{"complete": true}); err != nil {
		return "", err
	}
	return root, nil
}
func ensurePlainDirectory(path string) error {
	if info, err := os.Lstat(path); err == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("runtime directory must not be a symlink or file: %s", path)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.MkdirAll(path, 0700)
}
