package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

// Level represents the severity level of a log message
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel parses a string level to Level type
func ParseLevel(level string) Level {
	switch level {
	case "debug", "DEBUG":
		return LevelDebug
	case "info", "INFO":
		return LevelInfo
	case "warn", "WARN", "warning", "WARNING":
		return LevelWarn
	case "error", "ERROR":
		return LevelError
	case "fatal", "FATAL":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// Entry represents a single log entry
type Entry struct {
	Level     string                 `json:"level"`
	Timestamp string                 `json:"timestamp"`
	Message   string                 `json:"message"`
	Caller    string                 `json:"caller,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Logger is a structured JSON logger
type Logger struct {
	mu       sync.Mutex
	output   io.Writer
	minLevel Level
	fields   map[string]interface{}
}

// contextKey is the type for context keys
type contextKey string

const loggerKey contextKey = "logger"

// Global logger instance
var defaultLogger = New(os.Stdout, LevelInfo)

// New creates a new Logger instance
func New(output io.Writer, minLevel Level) *Logger {
	return &Logger{
		output:   output,
		minLevel: minLevel,
		fields:   make(map[string]interface{}),
	}
}

// SetDefault sets the default logger
func SetDefault(l *Logger) {
	defaultLogger = l
}

// Default returns the default logger
func Default() *Logger {
	return defaultLogger
}

// WithField returns a new logger with the given field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := &Logger{
		output:   l.output,
		minLevel: l.minLevel,
		fields:   make(map[string]interface{}),
	}
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	newLogger.fields[key] = value
	return newLogger
}

// WithFields returns a new logger with the given fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := &Logger{
		output:   l.output,
		minLevel: l.minLevel,
		fields:   make(map[string]interface{}),
	}
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return newLogger
}

// WithContext adds logger to context
func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext retrieves logger from context
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerKey).(*Logger); ok {
		return logger
	}
	return defaultLogger
}

// WithRequestID returns a new logger with request ID field
func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.WithField("request_id", requestID)
}

// log writes a log entry
func (l *Logger) log(level Level, msg string, args ...interface{}) {
	if level < l.minLevel {
		return
	}

	entry := Entry{
		Level:     level.String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Message:   fmt.Sprintf(msg, args...),
	}

	// Add caller information for error and fatal levels
	if level >= LevelError {
		if _, file, line, ok := runtime.Caller(2); ok {
			entry.Caller = fmt.Sprintf("%s:%d", file, line)
		}
	}

	// Add fields if any
	if len(l.fields) > 0 {
		entry.Fields = l.fields
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(l.output, `{"level":"ERROR","message":"failed to marshal log entry: %v"}`+"\n", err)
		return
	}

	l.output.Write(append(data, '\n'))
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(LevelDebug, msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(LevelInfo, msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(LevelWarn, msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(LevelError, msg, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.log(LevelFatal, msg, args...)
	os.Exit(1)
}

// Package-level functions for convenience

// Debug logs a debug message using the default logger
func Debug(msg string, args ...interface{}) {
	defaultLogger.Debug(msg, args...)
}

// Info logs an info message using the default logger
func Info(msg string, args ...interface{}) {
	defaultLogger.Info(msg, args...)
}

// Warn logs a warning message using the default logger
func Warn(msg string, args ...interface{}) {
	defaultLogger.Warn(msg, args...)
}

// Error logs an error message using the default logger
func Error(msg string, args ...interface{}) {
	defaultLogger.Error(msg, args...)
}

// Fatal logs a fatal message and exits using the default logger
func Fatal(msg string, args ...interface{}) {
	defaultLogger.Fatal(msg, args...)
}

// WithField returns a new logger with the given field using the default logger
func WithField(key string, value interface{}) *Logger {
	return defaultLogger.WithField(key, value)
}

// WithFields returns a new logger with the given fields using the default logger
func WithFields(fields map[string]interface{}) *Logger {
	return defaultLogger.WithFields(fields)
}
