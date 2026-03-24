//go:build gui

package gui_test

import (
	"fmt"

	"github.com/DementorAK/photometa/internal/adapter/gui"
	"github.com/DementorAK/photometa/internal/analyzer"
	"github.com/DementorAK/photometa/internal/platform/logger"
)

func ExampleNewGUI() {
	// This example demonstrates how to initialize and start the PhotoMeta GUI.
	// Note: Running a full GUI requires a graphical environment (X11, Wayland, or Windows/macOS).
	// In CI or headless environments, this example only simulates the setup.

	fmt.Println("Initializing dependencies...")

	// 1. Initialize the logger from infrastructure
	log := logger.NewSlogLogger()

	// 2. Create your service implementation (the "core")
	analyzer := analyzer.NewService(log)

	// 3. Initialize the GUI adapter with the service
	// We verify the function exists without calling ui.Start() to avoid headless crashes
	ui := gui.NewGUI(analyzer)

	fmt.Println("GUI initialized successfully.")

	// In a real application, you would call:
	// ui.Start() // This blocks until the window is closed

	if ui != nil {
		fmt.Println("Ready to show the main window.")
	}

	// Output:
	// Initializing dependencies...
	// GUI initialized successfully.
	// Ready to show the main window.
}
