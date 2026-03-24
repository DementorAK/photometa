package analyzer_test

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/DementorAK/photometa/internal/analyzer"
	"github.com/DementorAK/photometa/internal/platform/logger"
)

func ExampleService_AnalyzeFile() {
	// 1. Initialize dependencies (logger)
	// In a real application, you might use a customized slog logger.
	logger := logger.NewSlogLoggerWithLevel(slog.LevelInfo)

	// 2. Create the service
	service := analyzer.NewService(logger)

	// 3. Analyze a file
	// Note: Replace "example.jpg" with a real path to see actual metadata.
	ctx := context.Background()
	img, err := service.AnalyzeFile(ctx, "example.jpg")
	if err != nil {
		// For the sake of the example, we handle the error by printing it.
		// In a real app, use proper error handling.
		fmt.Printf("Analysis failed: %v\n", err)
		return
	}

	fmt.Printf("File: %s\n", img.Name)
	fmt.Printf("Format: %s\n", img.Metadata.Format)
}

func ExampleService_ScanDirectory() {
	logger := logger.NewSlogLogger()
	service := analyzer.NewService(logger)

	ctx := context.Background()
	images, err := service.ScanDirectory(ctx, "./photos")
	if err != nil {
		fmt.Printf("Scan failed: %v\n", err)
		return
	}

	fmt.Printf("Found %d images\n", len(images))
}
