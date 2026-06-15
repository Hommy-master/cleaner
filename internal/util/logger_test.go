package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitLogger(t *testing.T) {
	dir := t.TempDir()
	ResetLoggerForTest()

	if err := InitLogger(dir); err != nil {
		t.Fatalf("InitLogger: %v", err)
	}

	Logger().Print("test log line")

	logPath := filepath.Join(dir, logFileName)
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(data), "test log line") {
		t.Fatalf("log file content = %q, want test log line", string(data))
	}

	ResetLoggerForTest()
}

func TestInitLoggerInvalidDir(t *testing.T) {
	ResetLoggerForTest()

	invalidDir := filepath.Join(t.TempDir(), "missing", "nested")
	if err := InitLogger(invalidDir); err == nil {
		t.Fatal("expected error for invalid log directory")
	}
}
