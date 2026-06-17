package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const (
	logFileName   = "cleaner.log"
	logTimeLayout = "2006-01-02 15:04:05"
)

// Logger writes timestamped logs with source file and line number.
type Logger struct {
	mu      sync.Mutex
	out     io.Writer
	stdout  *os.File
	logFile *os.File
	debug   bool
}

var (
	logger     *Logger
	logFile    *os.File
	loggerOnce sync.Once
	loggerErr  error
)

// NewLogger creates a logger that writes to w.
func NewLogger(w io.Writer) *Logger {
	return &Logger{out: w}
}

// InitLogger configures logging to both stdout and cleaner.log in dir.
func InitLogger(dir string) error {
	loggerOnce.Do(func() {
		disableQuickEdit()

		logPath := filepath.Join(dir, logFileName)
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			loggerErr = fmt.Errorf("open log file %q: %w", logPath, err)
			return
		}
		logFile = f
		logger = &Logger{
			stdout:  os.Stdout,
			logFile: f,
		}
	})
	return loggerErr
}

// GetLogger returns the shared logger instance.
func GetLogger() *Logger {
	if logger == nil {
		return NewLogger(os.Stderr)
	}
	return logger
}

// SetDebug enables or disables debug-level log output.
func (l *Logger) SetDebug(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debug = enabled
}

// Printf writes a formatted log line (errors and deletion results).
func (l *Logger) Printf(format string, args ...any) {
	l.write(fmt.Sprintf(format, args...))
}

// Debugf writes a formatted log line only when debug mode is enabled.
func (l *Logger) Debugf(format string, args ...any) {
	l.mu.Lock()
	debug := l.debug
	l.mu.Unlock()
	if !debug {
		return
	}
	l.write(fmt.Sprintf(format, args...))
}

// Print writes a log line.
func (l *Logger) Print(v ...any) {
	l.write(fmt.Sprint(v...))
}

func (l *Logger) write(msg string) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	} else {
		file = filepath.Base(file)
	}
	lineText := fmt.Sprintf("%s %s:%d %s\n", time.Now().Format(logTimeLayout), file, line, msg)

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.stdout != nil || l.logFile != nil {
		if l.stdout != nil {
			_, _ = io.WriteString(l.stdout, lineText)
			flushConsole(l.stdout)
		}
		if l.logFile != nil {
			_, _ = io.WriteString(l.logFile, lineText)
			_ = l.logFile.Sync()
		}
		return
	}

	_, _ = io.WriteString(l.out, lineText)
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
