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

// IgnoreRule describes one resolved ignore entry under a cleanup root.
type IgnoreRule struct {
	RelPath string
	IsDir   bool
}

// BuildIgnoreRules resolves ignore entries relative to root against the filesystem.
func BuildIgnoreRules(root string, ignore []string) ([]IgnoreRule, error) {
	root = filepath.Clean(root)
	rules := make([]IgnoreRule, 0, len(ignore))

	for _, item := range ignore {
		rel := NormalizeRelativePath(item)
		if rel == "" {
			continue
		}

		fullPath := filepath.Join(root, rel)
		kind, err := ClassifyPath(fullPath)
		if err != nil {
			return nil, fmt.Errorf("classify ignore path %q: %w", rel, err)
		}

		rules = append(rules, IgnoreRule{
			RelPath: rel,
			IsDir:   kind == PathDirectory,
		})
	}

	return rules, nil
}

// IsIgnored reports whether relPath should be skipped according to ignore rules.
// Directory rules protect the directory and everything beneath it.
// File rules protect only the exact relative path.
func IsIgnored(relPath string, rules []IgnoreRule) bool {
	relPath = NormalizeRelativePath(relPath)
	if relPath == "" {
		return false
	}

	for _, rule := range rules {
		if rule.IsDir {
			if relPath == rule.RelPath || strings.HasPrefix(relPath, rule.RelPath+"/") {
				return true
			}
			continue
		}
		if relPath == rule.RelPath {
			return true
		}
	}
	return false
}
