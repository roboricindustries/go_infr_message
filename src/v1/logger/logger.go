package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// JSONFormatter defines your custom JSON log format
type JSONFormatter struct{}

// Format ensures logs match this JSON structure:
// {"time":"2025-01-22T12:00:00.000Z","level":"INFO","line":34,"msg":"Application started"}
func (f *JSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.UTC().Format(time.RFC3339Nano)
	level := strings.ToUpper(entry.Level.String())

	line := 0
	if entry.HasCaller() {
		line = entry.Caller.Line
	}

	msg := escapeString(entry.Message)
	logLine := fmt.Sprintf(
		`{"time":"%s","level":"%s","line":%d,"msg":"%s"}`+"\n",
		timestamp, level, line, msg,
	)
	return []byte(logLine), nil
}

func escapeString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

// ErrorHook is a custom Logrus hook that writes only error-level (and above) logs
// to a separate writer (e.g., a second file).
type ErrorHook struct {
	writer io.Writer
}

// Levels defines which log levels trigger this hook.
func (h *ErrorHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

// Fire is called when a log entry with level >= Error is emitted.
func (h *ErrorHook) Fire(entry *logrus.Entry) error {
	// Format the log entry using the logger's existing formatter
	lineBytes, err := entry.Logger.Formatter.Format(entry)
	if err != nil {
		return err
	}

	_, writeErr := h.writer.Write(lineBytes)
	return writeErr
}

// buildErrorFilename modifies "filename.log" → "filename_error.log".
func buildErrorFilename(filename string) string {
	ext := filepath.Ext(filename)             // e.g. ".log"
	base := strings.TrimSuffix(filename, ext) // e.g. "filename"
	return base + "_error" + ext              // e.g. "filename_error.log"
}

// NewLogger creates a Logrus logger that writes all logs to logFile
// and optionally writes error+ logs to a second file named filename_error.ext
func NewLogger(logDir, logFile, logLevel string, withErrorFile bool) *logrus.Logger {
	l := logrus.New()

	// Enable line numbers
	l.SetReportCaller(true)

	// Use custom JSON formatter
	l.SetFormatter(&JSONFormatter{})

	// Ensure the directory exists
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		panic(fmt.Sprintf("failed to create log directory %q: %v", logDir, err))
	}

	// Main log file
	mainLogPath := filepath.Join(logDir, logFile)
	mainLogger := &lumberjack.Logger{
		Filename:   mainLogPath,
		MaxSize:    10, // 10MB per file
		MaxBackups: 3,  // Keep 3 rotated files
		MaxAge:     30, // Keep logs for 30 days (adjustable)
		Compress:   true,
	}
	l.SetOutput(mainLogger)

	// If withErrorFile is true, create a second file for error+ logs
	if withErrorFile {
		errorFilename := buildErrorFilename(logFile) // e.g. "filename_error.log"
		errorLogPath := filepath.Join(logDir, errorFilename)
		errorLogger := &lumberjack.Logger{
			Filename:   errorLogPath,
			MaxSize:    10, // 10MB per file
			MaxBackups: 3,  // Keep 3 rotated files
			MaxAge:     30, // Keep logs for 30 days (adjustable)
			Compress:   true,
		}

		// Add the ErrorHook that writes only error+ logs to errorLogger
		l.AddHook(&ErrorHook{writer: errorLogger})
	}

	// Parse log level
	lvl, parseErr := logrus.ParseLevel(logLevel)
	if parseErr != nil {
		lvl = logrus.InfoLevel
	}
	l.SetLevel(lvl)

	return l
}
