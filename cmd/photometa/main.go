package main

import (
	"os"

	"github.com/DementorAK/photometa/internal/platform/logger"

	"github.com/DementorAK/photometa/internal/analyzer"

	"github.com/DementorAK/photometa/internal/adapter/cli"
)

func main() {
	// ============================================
	// COMPOSITION ROOT - All dependencies are wired here
	// ============================================

	// 1. Create infrastructure dependencies
	logger := logger.NewSlogLogger()

	// 2. Create application layer with dependencies
	service := analyzer.NewService(logger)

	// 3. Create adapter layer with dependencies
	cmd := cli.NewRootCmd(service, logger)

	// 4. Run the application
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
