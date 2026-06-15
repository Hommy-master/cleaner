//go:build windows

package util

import (
	"os"
	"syscall"
)

func flushConsole(f *os.File) {
	if f == nil {
		return
	}
	_ = syscall.FlushFileBuffers(syscall.Handle(f.Fd()))
}
