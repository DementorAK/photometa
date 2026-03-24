//go:build !gui

package gui

import (
	"fmt"
	"os"

	"github.com/DementorAK/photometa/internal/port"
)

type GUI struct {
	service port.ImageAnalyzer
}

func NewGUI(service port.ImageAnalyzer) *GUI {
	return &GUI{service: service}
}

func (g *GUI) Start() {
	fmt.Println("GUI mode is not available in this build (Requires GCC/CGo).")
	fmt.Println("Please install a C compiler and rebuild with '-tags gui' or use CLI mode:")
	fmt.Println("  photometa <dir>")
	fmt.Println("  photometa --server")
	os.Exit(1)
}
