package fake

import (
	"context"
	"io"

	"github.com/DementorAK/photometa/internal/domain"
)

// AnalyzeStreamCall represents the arguments of an AnalyzeStream call.
type AnalyzeStreamCall struct {
	Name string
	Size int64
}

// FakeImageAnalyzer implements port.ImageAnalyzer for testing.
type MockImageAnalyzer struct {
	// 1. Function-based faking (for complex per-test behavior)
	AnalyzeStreamFn func(ctx context.Context, r io.Reader, name string, size int64) (*domain.ImageFile, error)
	AnalyzeFileFn   func(ctx context.Context, path string) (*domain.ImageFile, error)
	ScanDirectoryFn func(ctx context.Context, root string) ([]domain.ImageFile, error)

	// 2. Result-based faking (for simpler static data/errors)
	AnalyzeStreamResult *domain.ImageFile
	AnalyzeStreamError  error
	AnalyzeFileResult   *domain.ImageFile
	AnalyzeFileError    error
	ScanDirectoryResult []domain.ImageFile
	ScanDirectoryError  error

	// Call tracking
	AnalyzeStreamCalls []AnalyzeStreamCall
	AnalyzeFileCalls   []string
	ScanDirectoryCalls []string
}

func NewMockImageAnalyzer() *MockImageAnalyzer {
	return &MockImageAnalyzer{
		// Default successful responses for result-based faking
		AnalyzeStreamResult: &domain.ImageFile{
			Name: "stream.jpg",
			Metadata: domain.Metadata{
				Format:   "jpeg",
				FileSize: 1024,
			},
		},
		AnalyzeFileResult: &domain.ImageFile{
			Name: "file.jpg",
			Metadata: domain.Metadata{
				Format:   "jpeg",
				FileSize: 2048,
			},
		},
		ScanDirectoryResult: []domain.ImageFile{},
	}
}

func (m *MockImageAnalyzer) AnalyzeStream(ctx context.Context, r io.Reader, name string, size int64) (*domain.ImageFile, error) {
	m.AnalyzeStreamCalls = append(m.AnalyzeStreamCalls, AnalyzeStreamCall{Name: name, Size: size})

	if m.AnalyzeStreamFn != nil {
		return m.AnalyzeStreamFn(ctx, r, name, size)
	}

	if m.AnalyzeStreamError != nil {
		return nil, m.AnalyzeStreamError
	}
	return m.AnalyzeStreamResult, nil
}

func (m *MockImageAnalyzer) AnalyzeFile(ctx context.Context, path string) (*domain.ImageFile, error) {
	m.AnalyzeFileCalls = append(m.AnalyzeFileCalls, path)

	if m.AnalyzeFileFn != nil {
		return m.AnalyzeFileFn(ctx, path)
	}

	if m.AnalyzeFileError != nil {
		return nil, m.AnalyzeFileError
	}

	// Return result with the actual path
	result := *m.AnalyzeFileResult
	result.Path = path
	result.Name = path
	return &result, nil
}

func (m *MockImageAnalyzer) ScanDirectory(ctx context.Context, root string) ([]domain.ImageFile, error) {
	m.ScanDirectoryCalls = append(m.ScanDirectoryCalls, root)

	if m.ScanDirectoryFn != nil {
		return m.ScanDirectoryFn(ctx, root)
	}

	if m.ScanDirectoryError != nil {
		return nil, m.ScanDirectoryError
	}
	return m.ScanDirectoryResult, nil
}
