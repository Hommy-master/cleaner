//go:build windows

package util

import (
	"os"
	"syscall"
	"time"
)

func fileCreatedAt(info os.FileInfo) time.Time {
	if sys, ok := info.Sys().(*syscall.Win32FileAttributeData); ok {
		return time.Unix(0, sys.CreationTime.Nanoseconds())
	}
	return info.ModTime()
}
