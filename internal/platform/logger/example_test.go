package logger_test

import (
	"log/slog"

	"github.com/DementorAK/photometa/internal/platform/logger"
)

func ExampleNewSlogLogger() {
	// Create a default logger (INFO level)
	log := logger.NewSlogLogger()

	log.Info("Application started")
	log.Debug("This will not be printed by default")

	// Output is written to stdout in logfmt format
}

func ExampleNewSlogLoggerWithLevel() {
	// Create a logger with DEBUG level
	log := logger.NewSlogLoggerWithLevel(slog.LevelDebug)

	log.Debug("Debug mode enabled")
}
