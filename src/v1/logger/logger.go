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

	// Build base JSON
	logLine := fmt.Sprintf(`{"time":"%s","level":"%s","line":%d,"msg":"%s"`,
		timestamp, level, line, msg,
	)

	// Include extra fields from entry.Data (like latency, client_ip, etc.)
	for k, v := range entry.Data {
		logLine += fmt.Sprintf(`,"%s":"%v"`, k, v)
	}

	// Close the JSON object
	logLine += "}\n"

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
