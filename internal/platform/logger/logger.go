package logger

import (
	"log/slog"
	"os"

	"github.com/DementorAK/photometa/internal/port"
)

// Ensure SlogLogger implements port.Logger
var _ port.Logger = (*SlogLogger)(nil)

// SlogLogger is a Logger implementation using Go's structured logging (slog).
type SlogLogger struct {
	logger *slog.Logger
}

// NewSlogLogger creates a new SlogLogger with default JSON handler.
func NewSlogLogger() *SlogLogger {
	return &SlogLogger{
		logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}
}

// NewSlogLoggerWithLevel creates a logger with custom level.
func NewSlogLoggerWithLevel(level slog.Level) *SlogLogger {
	return &SlogLogger{
		logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		})),
	}
}

func (l *SlogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *SlogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *SlogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *SlogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}
