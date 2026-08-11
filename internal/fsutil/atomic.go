// Package fsutil holds small filesystem helpers shared across packages.
package fsutil

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// WriteFile writes data to path atomically: it writes a temporary file in the
// same directory, then renames it over the target. A crash or a concurrent
// reader therefore sees either the old file or the new one, never a truncated
// mix of both. grapes reloads issues from a file watcher, so a half-written
// meta.toml would otherwise surface as a "malformed issue" for one refresh.

// AtomicFile describes one file in a multi-file atomic commit. All temporary
// files are fully written and synced before any destination is replaced.
type AtomicFile struct {
	Path string
	Data []byte
	Perm os.FileMode
}

// WriteFiles stages and commits several files. Filesystem renames cannot form
// a true transaction, but staging every file first ensures preparation errors
// leave all existing destinations untouched.
func WriteFiles(files []AtomicFile) error {
	if len(files) == 0 {
		return nil
	}
	type staged struct {
		AtomicFile
		tmpName        string
		backupName     string
		hadDestination bool
		renamed        bool
	}
	stagedFiles := make([]staged, 0, len(files))
	cleanup := func() {
		for _, f := range stagedFiles {
			_ = os.Remove(f.tmpName)
			_ = os.Remove(f.backupName)
		}
	}
	for _, file := range files {
		dir := filepath.Dir(file.Path)
		tmp, err := os.CreateTemp(dir, "."+filepath.Base(file.Path)+".tmp-*")
		if err != nil {
			cleanup()
			return fmt.Errorf("creating temp file in %s: %w", dir, err)
		}
		tmpName := tmp.Name()
		stagedFiles = append(stagedFiles, staged{AtomicFile: file, tmpName: tmpName})
		if _, err := tmp.Write(file.Data); err != nil {
			_ = tmp.Close()
			cleanup()
			return fmt.Errorf("writing %s: %w", tmpName, err)
		}
		if err := tmp.Chmod(file.Perm); err != nil {
			_ = tmp.Close()
			cleanup()
			return fmt.Errorf("chmod %s: %w", tmpName, err)
		}
		if err := tmp.Sync(); err != nil {
			_ = tmp.Close()
			cleanup()
			return fmt.Errorf("syncing %s: %w", tmpName, err)
		}
		if err := tmp.Close(); err != nil {
			cleanup()
			return fmt.Errorf("closing %s: %w", tmpName, err)
		}
	}
	for i := range stagedFiles {
		file := &stagedFiles[i]
		info, err := os.Stat(file.Path)
		switch {
		case err == nil && !info.Mode().IsRegular():
			cleanup()
			return fmt.Errorf("destination %s is not a regular file", file.Path)
		case err != nil && !os.IsNotExist(err):
			cleanup()
			return fmt.Errorf("checking destination %s: %w", file.Path, err)
		case err == nil:
			backup, backupErr := backupFile(file.Path, info.Mode().Perm())
			if backupErr != nil {
				cleanup()
				return backupErr
			}
			file.backupName = backup
			file.hadDestination = true
		}
	}

	rollback := func() error {
		var rollbackErrs []error
		for i := len(stagedFiles) - 1; i >= 0; i-- {
			file := &stagedFiles[i]
			if !file.renamed {
				continue
			}
			var err error
			if file.hadDestination {
				err = os.Rename(file.backupName, file.Path)
				if err == nil {
					file.backupName = ""
				}
			} else {
				err = os.Remove(file.Path)
			}
			if err != nil && !os.IsNotExist(err) {
				rollbackErrs = append(rollbackErrs, fmt.Errorf("restoring %s: %w", file.Path, err))
			}
			file.renamed = false
		}
		return errors.Join(rollbackErrs...)
	}

	for i := range stagedFiles {
		file := &stagedFiles[i]
		if err := os.Rename(file.tmpName, file.Path); err != nil {
			rollbackErr := rollback()
			cleanup()
			return errors.Join(fmt.Errorf("renaming %s to %s: %w", file.tmpName, file.Path, err), rollbackErr)
		}
		file.renamed = true
	}
	parents := make(map[string]struct{}, len(stagedFiles))
	for _, file := range stagedFiles {
		parents[filepath.Dir(file.Path)] = struct{}{}
	}
	for dir := range parents {
		if err := syncParent(dir); err != nil {
			rollbackErr := rollback()
			cleanup()
			return errors.Join(err, rollbackErr)
		}
	}
	cleanup()
	return nil
}

func backupFile(path string, perm os.FileMode) (string, error) {
	src, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening destination %s for backup: %w", path, err)
	}
	defer src.Close()
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".bak-*")
	if err != nil {
		return "", fmt.Errorf("creating backup for %s: %w", path, err)
	}
	tmpName := tmp.Name()
	cleanup := func(err error) (string, error) {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return "", err
	}
	if _, err := io.Copy(tmp, src); err != nil {
		return cleanup(fmt.Errorf("backing up %s: %w", path, err))
	}
	if err := tmp.Chmod(perm); err != nil {
		return cleanup(fmt.Errorf("chmod backup for %s: %w", path, err))
	}
	if err := tmp.Sync(); err != nil {
		return cleanup(fmt.Errorf("syncing backup for %s: %w", path, err))
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return "", fmt.Errorf("closing backup for %s: %w", path, err)
	}
	return tmpName, nil
}

func syncParent(dir string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	dirFile, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("opening parent directory %s after rename: %w", dir, err)
	}
	syncErr := dirFile.Sync()
	closeErr := dirFile.Close()
	if syncErr != nil {
		return fmt.Errorf("syncing parent directory %s: %w", dir, syncErr)
	}
	if closeErr != nil {
		return fmt.Errorf("closing parent directory %s: %w", dir, closeErr)
	}
	return nil
}

// WriteFile writes one file atomically.
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
	return syncParent(dir)
}
