package data

import (
	"golang.org/x/sys/windows"
)

func flockExclusive(fd uintptr) error {
	ol := new(windows.Overlapped)
	return windows.LockFileEx(windows.Handle(fd), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, ol)
}

func flockUnlock(fd uintptr) error {
	ol := new(windows.Overlapped)
	return windows.UnlockFileEx(windows.Handle(fd), 0, 1, 0, ol)
}
