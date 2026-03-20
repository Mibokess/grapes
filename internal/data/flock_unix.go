//go:build !windows

package data

import "syscall"

func flockExclusive(fd uintptr) error {
	return syscall.Flock(int(fd), syscall.LOCK_EX)
}

func flockUnlock(fd uintptr) error {
	return syscall.Flock(int(fd), syscall.LOCK_UN)
}
