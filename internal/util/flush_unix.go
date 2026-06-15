//go:build !windows

package util

import "os"

func flushConsole(f *os.File) {
	if f != nil {
		_ = f.Sync()
	}
}
