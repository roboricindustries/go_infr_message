package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// JSONFormatter defines your custom JSON log format
type JSONFormatter struct{}

func (f *JSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.UTC().Format(time.RFC3339Nano)
	level := strings.ToUpper(entry.Level.String())

	line := 0
	if entry.HasCaller() {
		line = entry.Caller.Line
	}

	msg := escapeString(entry.Message)

	// Example JSON structure:
	// {"time":"2025-01-22T12:00:00.000Z","level":"INFO","line":34,"msg":"Application started"}
	logLine := fmt.Sprintf(
		`{"time":"%s","level":"%s","line":%d,"msg":"%s"}`+"\n",
		timestamp, level, line, msg,
	)
	return []byte(logLine), nil
}

// escapeString ensures quotes in the message won't break JSON.
func escapeString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

// newLogger creates a Logrus logger that writes to the specified
// log file in logDir, with the given level ("debug", "info", etc.).
func NewLogger(logDir, logFile, logLevel string) *logrus.Logger {
	// 1. Create a new logger
	l := logrus.New()

	// 2. Enable line numbers
	l.SetReportCaller(true)

	// 3. Use custom JSON formatter
	l.SetFormatter(&JSONFormatter{})

	// 4. Ensure the directory exists
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		// If it fails to create the directory, we can fallback to default
		// or panic/return a default logger. Here, let's just panic for simplicity.
		panic(fmt.Sprintf("failed to create log directory %q: %v", logDir, err))
	}

	// 5. Open or create the log file
	logPath := filepath.Join(logDir, logFile)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		panic(fmt.Sprintf("failed to open log file %q: %v", logPath, err))
	}

	// 6. Direct log output to that file
	l.SetOutput(f)

	// 7. Parse and set log level (default to INFO if invalid)
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	l.SetLevel(lvl)

	return l
}
