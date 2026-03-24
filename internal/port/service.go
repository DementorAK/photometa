package port

import (
	"context"
	"io"

	"github.com/DementorAK/photometa/internal/domain"
)

// ImageAnalyzer defines the core operations for analyzing images.
// This is the primary port (use case interface) of the application.
// Adapters (CLI, HTTP, GUI) depend on this interface, not the concrete implementation.
type ImageAnalyzer interface {
	// AnalyzeStream processes an image from a generic reader (stdin, http request).
	// size can be 0 if unknown.
	AnalyzeStream(ctx context.Context, r io.Reader, name string, size int64) (*domain.ImageFile, error)

	// AnalyzeFile processes a single file path.
	AnalyzeFile(ctx context.Context, path string) (*domain.ImageFile, error)

	// ScanDirectory walks a directory and returns all found images.
	ScanDirectory(ctx context.Context, root string) ([]domain.ImageFile, error)
}

// Logger defines logging operations.
// Implementations should provide structured logging with the following levels:
//   - Info: Significant application events
//   - Warn: Recoverable issues that should be noted
//   - Error: Operation failures that may affect results
//   - Debug: Detailed information for troubleshooting
type Logger interface {
	// Info logs significant events (e.g., "analyzing file", "scan completed").
	Info(msg string, args ...any)
	// Error logs operation failures (e.g., "failed to read file").
	Error(msg string, args ...any)
	// Warn logs recoverable issues (e.g., "unhandled metadata tag", "partial scan").
	Warn(msg string, args ...any)
	// Debug logs detailed troubleshooting information.
	Debug(msg string, args ...any)
}
