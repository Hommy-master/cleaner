package cleaner

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"cleaner/internal/config"
	"cleaner/internal/util"
)

func testLogger(t *testing.T) (*util.Logger, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	return util.NewLogger(&buf), &buf
}

func TestCleanFileDeletesExistingFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "remove.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	logger, buf := testLogger(t)
	cfg := &config.Config{Files: []string{filePath}}
	New(cfg, logger).Run()

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatal("file should be deleted")
	}
	if !bytes.Contains(buf.Bytes(), []byte("deleted file: "+absPath)) {
		t.Fatalf("log = %q, want deleted absolute path", buf.String())
	}
}

func TestCleanFileSkipsMissing(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.txt")

	logger, buf := testLogger(t)
	cfg := &config.Config{Files: []string{missing}}
	New(cfg, logger).Run()

	if !bytes.Contains(buf.Bytes(), []byte("does not exist")) {
		t.Fatalf("log = %q, want missing file message", buf.String())
	}
}

func TestCleanFileSkipsDirectory(t *testing.T) {
	dir := t.TempDir()

	logger, buf := testLogger(t)
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

	removeAbs, err := filepath.Abs(removeFile)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	nestedRemoveAbs, err := filepath.Abs(nestedRemove)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	logger, buf := testLogger(t)
	cfg := &config.Config{
		Dirs: []config.DirConfig{
			{
				Path:   root,
				Ignore: []string{"filename", "nested/keep.exe"},
			},
		},
	}
	New(cfg, logger).Run()

	assertExists(t, keepFile)
	assertExists(t, nestedKeep)
	assertNotExists(t, removeFile)
	assertNotExists(t, nestedRemove)
	assertExists(t, nestedDir)
	assertExists(t, root)

	if !bytes.Contains(buf.Bytes(), []byte("deleted file: "+removeAbs)) {
		t.Fatalf("log = %q, want deleted root file absolute path", buf.String())
	}
	if !bytes.Contains(buf.Bytes(), []byte("deleted file: "+nestedRemoveAbs)) {
		t.Fatalf("log = %q, want deleted nested file absolute path", buf.String())
	}
}

func TestCleanDirSkipsMissingDirectory(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing")

	logger, buf := testLogger(t)
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

	logger, buf := testLogger(t)
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

	logger, _ := testLogger(t)
	New(cfgWithDir(root, "keepdir"), logger).Run()

	assertExists(t, keepFile)
	assertNotExists(t, otherFile)
	assertExists(t, keepDir)
}

func TestRemoveDirContentsEmptyDir(t *testing.T) {
	root := t.TempDir()
	logger, _ := testLogger(t)
	c := New(&config.Config{}, logger)
	if err := c.removeDirContents(root, nil); err != nil {
		t.Fatalf("removeDirContents: %v", err)
	}
	if _, err := os.Stat(root); err != nil {
		t.Fatal("root directory should remain")
	}
}

func TestCleanDirPreservesIgnoredVersionDirectory(t *testing.T) {
	root := t.TempDir()
	versionDir := filepath.Join(root, "8.9.0.13361")
	nestedFile := filepath.Join(versionDir, "app.dll")
	otherFile := filepath.Join(root, "other.txt")

	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, p := range []string{nestedFile, otherFile} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	logger, _ := testLogger(t)
	New(cfgWithDir(root, "8.9.0.13361"), logger).Run()

	assertExists(t, versionDir)
	assertExists(t, nestedFile)
	assertNotExists(t, otherFile)
}

func TestCleanDirPreservesIgnoredFileOnly(t *testing.T) {
	root := t.TempDir()
	keepFile := filepath.Join(root, "Configure.ini")
	keepSimilar := filepath.Join(root, "Configure.ini.bak")
	removeFile := filepath.Join(root, "remove.txt")

	for _, p := range []string{keepFile, keepSimilar, removeFile} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	logger, _ := testLogger(t)
	New(cfgWithDir(root, "Configure.ini"), logger).Run()

	assertExists(t, keepFile)
	assertNotExists(t, keepSimilar)
	assertNotExists(t, removeFile)
}

func TestCleanFileLogFormat(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "remove.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	logger, buf := testLogger(t)
	New(&config.Config{Files: []string{filePath}}, logger).Run()

	pattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} cleaner\.go:\d+ deleted file:`)
	if !pattern.Match(buf.Bytes()) {
		t.Fatalf("log format mismatch: %q", buf.String())
	}
}

func cfgWithDir(root, ignore string) *config.Config {
	return &config.Config{
		Dirs: []config.DirConfig{
			{Path: root, Ignore: []string{ignore}},
		},
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
