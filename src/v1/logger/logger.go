package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Global loggers map and sync lock
var (
	loggers       = make(map[string]*logrus.Logger) // Stores multiple named loggers
	defaultLogger *logrus.Logger                    // Default logger (when no name is given)
	loggersMu     sync.RWMutex                      // Ensures thread safety
	once          sync.Once                         // Ensures loggers are initialized only once
)

// JSONFormatter formats logs in JSON style
type JSONFormatter struct {
	LoggerName string
}

// Format ensures logs match this JSON structure:
// {"time":"2025-01-22T12:00:00.000Z","level":"INFO","logger":"example_logger","line":34,"msg":"Application started"}
func (jf *JSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.UTC().Format(time.RFC3339Nano)
	level := strings.ToUpper(entry.Level.String())
	msg := escapeString(entry.Message)

	lineNum := 0
	if entry.HasCaller() {
		lineNum = entry.Caller.Line
	}

	logLine := fmt.Sprintf(
		`{"time":"%s","level":"%s","logger":"%s","line":%d,"msg":"%s"}`+"\n",
		timestamp, level, jf.LoggerName, lineNum, msg,
	)
	return []byte(logLine), nil
}

// InitializeLogger sets up a named logger and stores it in `loggers`
func InitializeLogger(loggerName, levelString, logsDir string) error {
	loggersMu.Lock()
	defer loggersMu.Unlock()

	// If logger already exists, do nothing
	if _, exists := loggers[loggerName]; exists {
		return nil
	}

	// Ensure logsDir exists
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Create logger
	logger, err := newLogger(loggerName, levelString, logsDir)
	if err != nil {
		return err
	}

	loggers[loggerName] = logger
	return nil
}

// InitializeDefaultLogger sets up a default logger that is used when no logger is specified.
func InitializeDefaultLogger(levelString, logsDir string) error {
	var err error
	once.Do(func() {
		loggersMu.Lock()
		defer loggersMu.Unlock()

		// Ensure logsDir exists
		if err = os.MkdirAll(logsDir, 0o755); err != nil {
			err = fmt.Errorf("failed to create logs directory: %w", err)
			return
		}

		// Create default logger
		defaultLogger, err = newLogger("default", levelString, logsDir)
	})
	return err
}

// GetLogger retrieves a named logger or falls back to defaultLogger
func GetLogger(loggerName string) *logrus.Logger {
	loggersMu.RLock()
	defer loggersMu.RUnlock()

	if logger, exists := loggers[loggerName]; exists {
		return logger
	}
	if defaultLogger != nil {
		return defaultLogger
	}
	panic("Logger not initialized. Call InitializeLogger() or InitializeDefaultLogger() first.")
}

// newLogger is a helper function to create a logrus.Logger
func newLogger(loggerName, levelString, logsDir string) (*logrus.Logger, error) {
	logFilePath := filepath.Join(logsDir, loggerName+".log")

	// Create or append to the log file
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %q: %w", logFilePath, err)
	}

	logger := logrus.New()
	logger.SetOutput(file)
	logger.SetReportCaller(true) // Enables line numbers
	logger.SetFormatter(&JSONFormatter{LoggerName: loggerName})

	level, err := logrus.ParseLevel(levelString)
	if err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", levelString, err)
	}
	logger.SetLevel(level)

	return logger, nil
}

// escapeString handles JSON escaping
func escapeString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

// -----------------------------------------------------------------------------
// Logging functions with designated loggers

func Info(loggerName, msg string) {
	GetLogger(loggerName).Info(msg)
}

func Debug(loggerName, msg string) {
	GetLogger(loggerName).Debug(msg)
}

func Warn(loggerName, msg string) {
	GetLogger(loggerName).Warn(msg)
}

func Error(loggerName, msg string) {
	GetLogger(loggerName).Error(msg)
}

func Infof(loggerName, format string, args ...interface{}) {
	GetLogger(loggerName).Infof(format, args...)
}

func Debugf(loggerName, format string, args ...interface{}) {
	GetLogger(loggerName).Debugf(format, args...)
}

func Warnf(loggerName, format string, args ...interface{}) {
	GetLogger(loggerName).Warnf(format, args...)
}

func Errorf(loggerName, format string, args ...interface{}) {
	GetLogger(loggerName).Errorf(format, args...)
}

// -----------------------------------------------------------------------------
// Logging functions for default logger (no loggerName needed)

func DefaultInfo(msg string) {
	GetLogger("default").Info(msg)
}

func DefaultDebug(msg string) {
	GetLogger("default").Debug(msg)
}

func DefaultWarn(msg string) {
	GetLogger("default").Warn(msg)
}

func DefaultError(msg string) {
	GetLogger("default").Error(msg)
}

func DefaultInfof(format string, args ...interface{}) {
	GetLogger("default").Infof(format, args...)
}

func DefaultDebugf(format string, args ...interface{}) {
	GetLogger("default").Debugf(format, args...)
}

func DefaultWarnf(format string, args ...interface{}) {
	GetLogger("default").Warnf(format, args...)
}

func DefaultErrorf(format string, args ...interface{}) {
	GetLogger("default").Errorf(format, args...)
}
