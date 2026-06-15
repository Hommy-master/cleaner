package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func absPath(t *testing.T, parts ...string) string {
	t.Helper()
	p := filepath.Join(parts...)
	if filepath.IsAbs(p) {
		return p
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(`D:\`, filepath.Join(parts...))
	}
	return filepath.Join("/", filepath.Join(parts...))
}

func TestLoadDefaultInterval(t *testing.T) {
	path := writeConfigFile(t, `{"dirs":[],"files":[]}`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Interval != defaultIntervalSeconds {
		t.Fatalf("Interval = %d, want %d", cfg.Interval, defaultIntervalSeconds)
	}
	if cfg.IntervalDuration() != 60*time.Second {
		t.Fatalf("IntervalDuration = %v, want 60s", cfg.IntervalDuration())
	}
}

func TestLoadValidConfig(t *testing.T) {
	filePath := absPath(t, "apps", "foo.txt")
	content := `{
		"interval": 10,
		"dirs": [{"path": "` + filepath.ToSlash(absPath(t, "apps", "JianyingPro")) + `", "ignore": ["filename"]}],
		"files": ["` + filepath.ToSlash(filePath) + `"]
	}`
	cfg, err := Load(writeConfigFile(t, content))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Interval != 10 {
		t.Fatalf("Interval = %d, want 10", cfg.Interval)
	}
	if len(cfg.Dirs) != 1 || len(cfg.Files) != 1 {
		t.Fatalf("unexpected dirs/files length: %+v", cfg)
	}
}

func TestValidateEmptyDirPath(t *testing.T) {
	cfg := &Config{Dirs: []DirConfig{{Path: ""}}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for empty dir path")
	}
}

func TestValidateRelativeFilePath(t *testing.T) {
	cfg := &Config{Files: []string{"relative/path.txt"}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for relative file path")
	}
}

func TestValidateEmptyFilePath(t *testing.T) {
	cfg := &Config{Files: []string{""}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for empty file path")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	_, err := Load(writeConfigFile(t, `{invalid`))
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("expected read error")
	}
}
