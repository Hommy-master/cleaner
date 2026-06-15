package cleaner

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"testing"

	"cleaner/internal/config"
)

func testLogger(t *testing.T) *log.Logger {
	t.Helper()
	var buf bytes.Buffer
	return log.New(&buf, "", 0)
}

func TestCleanFileDeletesExistingFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "remove.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cfg := &config.Config{Files: []string{filePath}}
	New(cfg, testLogger(t)).Run()

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatal("file should be deleted")
	}
}

func TestCleanFileSkipsMissing(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.txt")

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	cfg := &config.Config{Files: []string{missing}}
	New(cfg, logger).Run()

	if !bytes.Contains(buf.Bytes(), []byte("does not exist")) {
		t.Fatalf("log = %q, want missing file message", buf.String())
	}
}

func TestCleanFileSkipsDirectory(t *testing.T) {
	dir := t.TempDir()

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	cfg := &config.Config{Files: []string{dir}}
	New(cfg, logger).Run()

	if !bytes.Contains(buf.Bytes(), []byte("is a directory")) {
		t.Fatalf("log = %q, want directory skip message", buf.String())
	}
}

func TestCleanDirRemovesContentsExceptIgnored(t *testing.T) {
	root := t.TempDir()

	keepFile := filepath.Join(root, "filename")
	removeFile := filepath.Join(root, "remove.txt")
	nestedDir := filepath.Join(root, "nested")
	nestedKeep := filepath.Join(nestedDir, "keep.exe")
	nestedRemove := filepath.Join(nestedDir, "drop.txt")

	for _, p := range []struct {
		path    string
		content string
	}{
		{keepFile, "keep"},
		{removeFile, "remove"},
		{nestedKeep, "keep"},
		{nestedRemove, "remove"},
	} {
		if err := os.MkdirAll(filepath.Dir(p.path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(p.path, []byte(p.content), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	cfg := &config.Config{
		Dirs: []config.DirConfig{
			{
				Path:   root,
				Ignore: []string{"filename", "nested/keep.exe"},
			},
		},
	}
	New(cfg, testLogger(t)).Run()

	assertExists(t, keepFile)
	assertExists(t, nestedKeep)
	assertNotExists(t, removeFile)
	assertNotExists(t, nestedRemove)
	assertExists(t, nestedDir)
	assertExists(t, root)
}

func TestCleanDirSkipsMissingDirectory(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing")

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	cfg := &config.Config{
		Dirs: []config.DirConfig{{Path: missing}},
	}
	New(cfg, logger).Run()

	if !bytes.Contains(buf.Bytes(), []byte("does not exist")) {
		t.Fatalf("log = %q, want missing directory message", buf.String())
	}
}

func TestCleanDirSkipsFilePath(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "notdir.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	cfg := &config.Config{
		Dirs: []config.DirConfig{{Path: filePath}},
	}
	New(cfg, logger).Run()

	if !bytes.Contains(buf.Bytes(), []byte("is a file")) {
		t.Fatalf("log = %q, want file skip message", buf.String())
	}
	if _, err := os.Stat(filePath); err != nil {
		t.Fatal("file path should remain when configured as directory")
	}
}

func TestCleanDirPreservesIgnoredDirectoryTree(t *testing.T) {
	root := t.TempDir()
	keepDir := filepath.Join(root, "keepdir")
	keepFile := filepath.Join(keepDir, "a.txt")
	otherFile := filepath.Join(root, "other.txt")

	if err := os.MkdirAll(keepDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, p := range []string{keepFile, otherFile} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	cfg := &config.Config{
		Dirs: []config.DirConfig{
			{Path: root, Ignore: []string{"keepdir"}},
		},
	}
	New(cfg, testLogger(t)).Run()

	assertExists(t, keepFile)
	assertNotExists(t, otherFile)
	assertExists(t, keepDir)
}

func TestRemoveDirContentsEmptyDir(t *testing.T) {
	root := t.TempDir()
	if err := removeDirContents(root, nil); err != nil {
		t.Fatalf("removeDirContents: %v", err)
	}
	if _, err := os.Stat(root); err != nil {
		t.Fatal("root directory should remain")
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %q to exist: %v", path, err)
	}
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected %q to not exist", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %q: %v", path, err)
	}
}
