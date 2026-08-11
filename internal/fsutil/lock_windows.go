//go:build windows

package fsutil

import "golang.org/x/sys/windows"

func lockExclusive(fd uintptr) error {
	overlapped := new(windows.Overlapped)
	return windows.LockFileEx(windows.Handle(fd), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, overlapped)
}

func unlock(fd uintptr) error {
	overlapped := new(windows.Overlapped)
	return windows.UnlockFileEx(windows.Handle(fd), 0, 1, 0, overlapped)
}
