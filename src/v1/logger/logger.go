// File: unilog.go
package unilog

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// JSONFormatter is a custom Logrus formatter that outputs JSON logs
// in the specific structure you requested.
type JSONFormatter struct {
	LoggerName string
}

// Format builds the log output in JSON form, e.g.:
//
//	{
//	  "time": "2025-01-22T12:00:00.000Z",
//	  "level": "INFO",
//	  "logger": "example_logger",
//	  "line": 34,
//	  "msg": "Application started"
//	}
func (jf *JSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Default fields
	level := strings.ToUpper(entry.Level.String())
	msg := entry.Message
	timestamp := entry.Time.UTC().Format(time.RFC3339Nano)

	// Extract caller line number if available
	var lineNum int
	if entry.HasCaller() {
		lineNum = entry.Caller.Line
	}

	// Build your desired JSON structure manually
	logLine := fmt.Sprintf(`{"time":"%s","level":"%s","logger":"%s","line":%d,"msg":"%s"}`+"\n",
		timestamp, level, jf.LoggerName, lineNum, escapeString(msg),
	)

	return []byte(logLine), nil
}

// Logger wraps a *logrus.Logger with your custom setup.
type Logger struct {
	*logrus.Logger
	name string
}

// NewLogger creates a new logger that writes JSON logs to a file
// and includes your custom JSON format.
//
//	name:         "example_logger" (appears in the "logger" field)
//	levelString:  e.g. "info", "debug", "warn", "error", ...
//	filePath:     path to the output log file
func NewLogger(name, levelString, filePath string) (*Logger, error) {
	log := logrus.New()

	// Parse log level
	level, err := logrus.ParseLevel(levelString)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	log.SetLevel(level)

	// Create or append to log file
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", filePath, err)
	}

	// Tell logrus to write to file
	log.SetOutput(file)

	// Enable Caller info so we know the code line
	log.SetReportCaller(true)

	// Attach the custom JSON formatter
	log.SetFormatter(&JSONFormatter{
		LoggerName: name,
	})

	return &Logger{Logger: log, name: name}, nil
}

// escapeString escapes quotation marks to keep JSON valid
func escapeString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

// Info logs an INFO level message
func (l *Logger) Info(msg string) {
	l.Logger.Info(msg)
}

// Debug logs a DEBUG level message
func (l *Logger) Debug(msg string) {
	l.Logger.Debug(msg)
}

// Error logs an ERROR level message
func (l *Logger) Error(msg string) {
	l.Logger.Error(msg)
}
