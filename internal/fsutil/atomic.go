// Package fsutil holds small filesystem helpers shared across packages.
package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile writes data to path atomically: it writes a temporary file in the
// same directory, then renames it over the target. A crash or a concurrent
// reader therefore sees either the old file or the new one, never a truncated
// mix of both. grapes reloads issues from a file watcher, so a half-written
// meta.toml would otherwise surface as a "malformed issue" for one refresh.
func WriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()

	// From here on every failure must remove the temp file.
	cleanup := func(err error) error {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}

	if _, err := tmp.Write(data); err != nil {
		return cleanup(fmt.Errorf("writing %s: %w", tmpName, err))
	}
	if err := tmp.Chmod(perm); err != nil {
		return cleanup(fmt.Errorf("chmod %s: %w", tmpName, err))
	}
	if err := tmp.Sync(); err != nil {
		return cleanup(fmt.Errorf("syncing %s: %w", tmpName, err))
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing %s: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming %s to %s: %w", tmpName, path, err)
	}
	return nil
}
