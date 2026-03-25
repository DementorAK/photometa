package port_test

import (
	"context"
	"fmt"
	"io"

	"github.com/DementorAK/photometa/internal/domain"
	"github.com/DementorAK/photometa/internal/port"
)

// MyImageService is an example implementation of the ImageAnalyzer interface.
type MyImageService struct{}

// AnalyzeStream implements port.ImageAnalyzer.
func (s *MyImageService) AnalyzeStream(_ context.Context, _ io.Reader, name string, _ int64) (*domain.ImageFile, error) {
	return &domain.ImageFile{Name: name}, nil
}

// AnalyzeFile implements port.ImageAnalyzer.
func (s *MyImageService) AnalyzeFile(_ context.Context, path string) (*domain.ImageFile, error) {
	return &domain.ImageFile{Path: path}, nil
}

// ScanDirectory implements port.ImageAnalyzer.
func (s *MyImageService) ScanDirectory(_ context.Context, _ string) ([]domain.ImageFile, error) {
	return []domain.ImageFile{}, nil
}

// Ensure MyImageService implements the interface at compile time.
var _ port.ImageAnalyzer = (*MyImageService)(nil)

func ExampleImageAnalyzer() {
	// This example demonstrates how to implement and use the ImageAnalyzer interface.
	// You might implement this interface to support a new image storage backend
	// or a specialized analysis engine.

	// 1. Create an instance of your implementation
	var service port.ImageAnalyzer = &MyImageService{}

	// 2. Use the service to analyze a file
	ctx := context.Background()
	img, err := service.AnalyzeFile(ctx, "/images/summer_trip.jpg")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Successfully analyzed: %s\n", img.Path)

	// Output:
	// Successfully analyzed: /images/summer_trip.jpg
}
