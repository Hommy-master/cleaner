package util

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	minLinuxPathLen   = 2
	minWindowsPathLen = 4
)

// IsAbsolutePath reports whether path is an absolute path on the current OS.
func IsAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

// ValidateDirPath checks that a directory path is non-empty and meets minimum length.
func ValidateDirPath(path string) error {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	minLen := minLinuxPathLen
	if runtime.GOOS == "windows" {
		minLen = minWindowsPathLen
	}
	if len(trimmed) < minLen {
		return fmt.Errorf("directory path %q is too short", trimmed)
	}
	return nil
}

// PathKind describes whether a path is missing, a file, or a directory.
type PathKind int

const (
	PathMissing PathKind = iota
	PathFile
	PathDirectory
)

// ClassifyPath returns the kind of path on disk.
func ClassifyPath(path string) (PathKind, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return PathMissing, nil
		}
		return PathMissing, err
	}
	if info.IsDir() {
		return PathDirectory, nil
	}
	return PathFile, nil
}

// NormalizeRelativePath converts a relative path to slash-separated form for comparison.
func NormalizeRelativePath(rel string) string {
	rel = filepath.Clean(rel)
	if rel == "." {
		return ""
	}
	return filepath.ToSlash(rel)
}

// IsIgnored reports whether relPath should be skipped according to ignore rules.
// Ignore entries are relative paths from the configured directory root.
func IsIgnored(relPath string, ignore []string) bool {
	relPath = NormalizeRelativePath(relPath)
	if relPath == "" {
		return false
	}

	for _, item := range ignore {
		item = NormalizeRelativePath(item)
		if item == "" {
			continue
		}
		if relPath == item {
			return true
		}
		if strings.HasPrefix(relPath, item+"/") {
			return true
		}
	}
	return false
}
