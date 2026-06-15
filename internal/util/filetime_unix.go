//go:build !windows

package util

import (
	"os"
	"time"
)

func fileCreatedAt(info os.FileInfo) time.Time {
	return info.ModTime()
}
