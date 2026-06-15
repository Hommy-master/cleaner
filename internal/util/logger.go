package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const logFileName = "cleaner.log"

var (
	logger     *log.Logger
	logFile    *os.File
	loggerOnce sync.Once
	loggerErr  error
)

// InitLogger configures logging to both stdout and cleaner.log in dir.
func InitLogger(dir string) error {
	loggerOnce.Do(func() {
		logPath := filepath.Join(dir, logFileName)
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			loggerErr = fmt.Errorf("open log file %q: %w", logPath, err)
			return
		}
		logFile = f
		w := io.MultiWriter(os.Stdout, f)
		logger = log.New(w, "", log.LstdFlags)
	})
	return loggerErr
}

// Logger returns the shared logger instance.
func Logger() *log.Logger {
	if logger == nil {
		return log.Default()
	}
	return logger
}

// ResetLoggerForTest resets the logger singleton for unit tests.
func ResetLoggerForTest() {
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
	logger = nil
	loggerOnce = sync.Once{}
	loggerErr = nil
}
