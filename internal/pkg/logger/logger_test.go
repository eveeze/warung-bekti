package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/eveeze/warung-backend/internal/pkg/logger"
)

func TestLogger(t *testing.T) {
	// ParseLevel
	if logger.ParseLevel("debug") != logger.LevelDebug {
		t.Errorf("expected LevelDebug")
	}
	if logger.ParseLevel("warn") != logger.LevelWarn {
		t.Errorf("expected LevelWarn")
	}
	if logger.ParseLevel("error") != logger.LevelError {
		t.Errorf("expected LevelError")
	}
	if logger.ParseLevel("fatal") != logger.LevelFatal {
		t.Errorf("expected LevelFatal")
	}
	if logger.ParseLevel("unknown") != logger.LevelInfo {
		t.Errorf("expected default LevelInfo")
	}

	// Level.String()
	if logger.LevelDebug.String() != "DEBUG" {
		t.Errorf("expected DEBUG")
	}
	if logger.LevelInfo.String() != "INFO" {
		t.Errorf("expected INFO")
	}
	if logger.LevelWarn.String() != "WARN" {
		t.Errorf("expected WARN")
	}
	if logger.LevelError.String() != "ERROR" {
		t.Errorf("expected ERROR")
	}
	if logger.LevelFatal.String() != "FATAL" {
		t.Errorf("expected FATAL")
	}
	if logger.Level(99).String() != "UNKNOWN" {
		t.Errorf("expected UNKNOWN")
	}

	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelDebug)

	l.Debug("debug message")
	l.Info("info message")
	l.Warn("warn message")
	l.Error("error message")

	// Logger with fields
	lf := l.WithField("user_id", "123").WithFields(map[string]interface{}{"role": "admin"}).WithRequestID("req-999")
	lf.Info("message with fields")

	// Context logger
	ctx := l.WithContext(context.Background())
	fromCtx := logger.FromContext(ctx)
	if fromCtx == nil {
		t.Errorf("expected logger from context")
	}
	fromEmpty := logger.FromContext(context.Background())
	if fromEmpty == nil {
		t.Errorf("expected default logger from empty context")
	}

	// Global log helpers
	logger.SetDefault(l)
	if logger.Default() != l {
		t.Errorf("expected default logger to match")
	}

	logger.Debug("global debug")
	logger.Info("global info")
	logger.Warn("global warn")
	logger.Error("global error")
	logger.WithField("k", "v").Info("global with field")
	logger.WithFields(map[string]interface{}{"k": "v"}).Info("global with fields")

	// Check output
	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	if len(lines) == 0 {
		t.Fatalf("expected logged lines, got 0")
	}

	var entry logger.Entry
	if err := json.Unmarshal(lines[0], &entry); err != nil {
		t.Fatalf("failed to parse log json: %v", err)
	}
	if entry.Message != "debug message" {
		t.Errorf("expected 'debug message', got '%s'", entry.Message)
	}
}
