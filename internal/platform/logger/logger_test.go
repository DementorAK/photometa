package logger

import (
	"bytes"
	"log/slog"
	"testing"
)

func TestNewSlogLogger(t *testing.T) {
	logger := NewSlogLogger()
	if logger == nil {
		t.Fatal("expected logger to be created")
	}
	if logger.logger == nil {
		t.Error("expected internal logger to be initialized")
	}
}

func TestNewSlogLoggerWithLevel(t *testing.T) {
	tests := []struct {
		name  string
		level slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewSlogLoggerWithLevel(tt.level)
			if logger == nil {
				t.Fatal("expected logger to be created")
			}
		})
	}
}

func TestSlogLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := &SlogLogger{
		logger: slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}

	logger.Info("test message", "key", "value")

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}
}

func TestSlogLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := &SlogLogger{
		logger: slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelError,
		})),
	}

	logger.Error("error occurred", "code", 500)

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}
}

func TestSlogLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	logger := &SlogLogger{
		logger: slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelWarn,
		})),
	}

	logger.Warn("warning message", "count", 3)

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}
}

func TestSlogLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := &SlogLogger{
		logger: slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	logger.Debug("debug message", "data", "test")

	output := buf.String()
	if output == "" {
		t.Error("expected log output at debug level")
	}
}

func TestSlogLogger_ImplementsPortLogger(_ *testing.T) {
	var _ interface {
		Info(msg string, args ...any)
		Error(msg string, args ...any)
		Warn(msg string, args ...any)
		Debug(msg string, args ...any)
	} = &SlogLogger{}
}
