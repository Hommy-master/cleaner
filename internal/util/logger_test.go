package util

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestInitLogger(t *testing.T) {
	dir := t.TempDir()
	ResetLoggerForTest()

	if err := InitLogger(dir); err != nil {
		t.Fatalf("InitLogger: %v", err)
	}

	GetLogger().Print("test log line")

	logPath := filepath.Join(dir, logFileName)
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "test log line") {
		t.Fatalf("log file content = %q, want test log line", content)
	}

	pattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} logger_test\.go:\d+ test log line\n$`)
	if !pattern.MatchString(content) {
		t.Fatalf("log format mismatch: %q", content)
	}

	ResetLoggerForTest()
}

func TestLoggerFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)
	logger.Printf("hello %s", "world")

	line := buf.String()
	pattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} logger_test\.go:\d+ hello world\n$`)
	if !pattern.MatchString(line) {
		t.Fatalf("log format mismatch: %q", line)
	}
}

func TestLoggerDebugf(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	logger.Debugf("debug only")
	if buf.Len() != 0 {
		t.Fatalf("Debugf without debug = %q, want empty", buf.String())
	}

	logger.SetDebug(true)
	logger.Debugf("debug enabled")
	if !strings.Contains(buf.String(), "debug enabled") {
		t.Fatalf("Debugf with debug = %q, want debug enabled", buf.String())
	}

	buf.Reset()
	logger.Printf("always logged")
	if !strings.Contains(buf.String(), "always logged") {
		t.Fatalf("Printf = %q, want always logged", buf.String())
	}
}

func TestInitLoggerInvalidDir(t *testing.T) {
	ResetLoggerForTest()

	invalidDir := filepath.Join(t.TempDir(), "missing", "nested")
	if err := InitLogger(invalidDir); err == nil {
		t.Fatal("expected error for invalid log directory")
	}
}
