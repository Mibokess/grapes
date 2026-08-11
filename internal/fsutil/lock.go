package fsutil

import "os"

// FileLock is a process-shared advisory lock backed by a persistent file.
// The file is intentionally not removed: unlinking a held lock lets another
// process lock a different inode and defeats exclusion.
type FileLock struct {
	file *os.File
}

// OpenFileLock opens or creates the lock file without acquiring it.
func OpenFileLock(path string, perm os.FileMode) (*FileLock, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, perm)
	if err != nil {
		return nil, err
	}
	return &FileLock{file: file}, nil
}

// Lock acquires the lock, waiting until another holder releases it.
func (l *FileLock) Lock() error {
	return lockExclusive(l.file.Fd())
}

// Unlock releases the lock.
func (l *FileLock) Unlock() error {
	return unlock(l.file.Fd())
}

// Close closes the lock file. Callers should unlock before closing.
func (l *FileLock) Close() error {
	return l.file.Close()
}
