package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DementorAK/photometa/internal/fake"
)

// ============================================================================
// SERVICE TESTS
// ============================================================================

func TestNewService(t *testing.T) {
	// Arrange
	mockLogger := fake.NewMockLogger()

	// Act
	service := NewService(mockLogger)

	// Assert
	if service == nil {
		t.Fatal("expected service to be created, got nil")
	}
	if service.logger == nil {
		t.Fatal("expected logger to be injected, got nil")
	}
}

func TestAnalyzeFile_ValidJPEG(t *testing.T) {
	// Arrange
	mockLogger := fake.NewMockLogger()
	service := NewService(mockLogger)

	// Create a temporary test JPEG file
	// Note: This is a minimal valid JPEG (just header markers)
	testFile := createTempFile(t, "test*.jpg", minimalJPEG())
	t.Cleanup(func() { _ = os.Remove(testFile) })

	// Act
	result, err := service.AnalyzeFile(context.Background(), testFile)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.Name != filepath.Base(testFile) {
		t.Errorf("expected name %q, got %q", filepath.Base(testFile), result.Name)
	}

	if result.Metadata.Format != "jpeg" {
		t.Errorf("expected format 'jpeg', got %q", result.Metadata.Format)
	}

	if result.Metadata.FileSize == 0 {
		t.Error("expected non-zero file size")
	}
}

func TestAnalyzeFile_NonExistentFile(t *testing.T) {
	// Arrange
	mockLogger := fake.NewMockLogger()
	service := NewService(mockLogger)

	// Act
	result, err := service.AnalyzeFile(context.Background(), "/non/existent/file.jpg")

	// Assert
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}

	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}

func TestAnalyzeFile_Directory(t *testing.T) {
	// Arrange
	mockLogger := fake.NewMockLogger()
	service := NewService(mockLogger)

	tempDir := t.TempDir()

	// Act
	result, err := service.AnalyzeFile(context.Background(), tempDir)

	// Assert
	if err == nil {
		t.Fatal("expected error for directory, got nil")
	}

	if !strings.Contains(err.Error(), "directory") {
		t.Errorf("expected error message to contain 'directory', got: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}

func TestScanDirectory_Empty(t *testing.T) {
	// Arrange
	mockLogger := fake.NewMockLogger()
	service := NewService(mockLogger)

	tempDir := t.TempDir()

	// Act
	results, err := service.ScanDirectory(context.Background(), tempDir)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results for empty directory, got %d", len(results))
	}
}

func TestScanDirectory_WithImages(t *testing.T) {
	// Arrange
	mockLogger := fake.NewMockLogger()
	service := NewService(mockLogger)

	tempDir := t.TempDir()

	// Create test files
	createFileInDir(t, tempDir, "photo1.jpg", minimalJPEG())
	createFileInDir(t, tempDir, "photo2.jpeg", minimalJPEG())
	createFileInDir(t, tempDir, "document.pdf", []byte("%PDF-1.4")) // Should be ignored

	// Act
	results, err := service.ScanDirectory(context.Background(), tempDir)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 images, got %d", len(results))
	}

	// Verify all results are images
	for _, img := range results {
		if img.Metadata.Format != "jpeg" {
			t.Errorf("expected format 'jpeg', got %q for %s", img.Metadata.Format, img.Name)
		}
	}
}

func TestScanDirectory_Nested(t *testing.T) {
	// Arrange
	mockLogger := fake.NewMockLogger()
	service := NewService(mockLogger)

	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Create files in root and subdirectory
	createFileInDir(t, tempDir, "root.jpg", minimalJPEG())
	createFileInDir(t, subDir, "nested.jpg", minimalJPEG())

	// Act
	results, err := service.ScanDirectory(context.Background(), tempDir)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 images (including nested), got %d", len(results))
	}
}

func TestAnalyzeStream(t *testing.T) {
	// Arrange
	mockLogger := fake.NewMockLogger()
	service := NewService(mockLogger)

	data := minimalJPEG()
	reader := strings.NewReader(string(data))

	// Act
	result, err := service.AnalyzeStream(context.Background(), reader, "uploaded.jpg", int64(len(data)))

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "uploaded.jpg" {
		t.Errorf("expected name 'uploaded.jpg', got %q", result.Name)
	}

	if result.Metadata.FileSize != int64(len(data)) {
		t.Errorf("expected size %d, got %d", len(data), result.Metadata.FileSize)
	}
}

// ============================================================================
// TEST HELPERS
// ============================================================================

// minimalJPEG returns bytes of a minimal valid JPEG file.
// This is the smallest valid JPEG that can be parsed.
func minimalJPEG() []byte {
	// Minimal JPEG: SOI + APP0 + DQT + SOF0 + DHT + SOS + EOI
	// This is a 1x1 pixel gray JPEG
	return []byte{
		0xFF, 0xD8, // SOI (Start of Image)
		0xFF, 0xE0, 0x00, 0x10, // APP0 marker
		0x4A, 0x46, 0x49, 0x46, 0x00, // "JFIF\0"
		0x01, 0x01, // Version 1.1
		0x00,       // Aspect ratio units (0 = no units)
		0x00, 0x01, // X density
		0x00, 0x01, // Y density
		0x00, 0x00, // Thumbnail dimensions (none)
		0xFF, 0xDB, 0x00, 0x43, 0x00, // DQT marker
		0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07,
		0x07, 0x07, 0x09, 0x09, 0x08, 0x0A, 0x0C, 0x14,
		0x0D, 0x0C, 0x0B, 0x0B, 0x0C, 0x19, 0x12, 0x13,
		0x0F, 0x14, 0x1D, 0x1A, 0x1F, 0x1E, 0x1D, 0x1A,
		0x1C, 0x1C, 0x20, 0x24, 0x2E, 0x27, 0x20, 0x22,
		0x2C, 0x23, 0x1C, 0x1C, 0x28, 0x37, 0x29, 0x2C,
		0x30, 0x31, 0x34, 0x34, 0x34, 0x1F, 0x27, 0x39,
		0x3D, 0x38, 0x32, 0x3C, 0x2E, 0x33, 0x34, 0x32,
		0xFF, 0xC0, 0x00, 0x0B, 0x08, // SOF0 marker
		0x00, 0x01, // Height = 1
		0x00, 0x01, // Width = 1
		0x01,             // Number of components = 1 (grayscale)
		0x01, 0x11, 0x00, // Component 1
		0xFF, 0xC4, 0x00, 0x1F, 0x00, // DHT marker (DC)
		0x00, 0x01, 0x05, 0x01, 0x01, 0x01, 0x01, 0x01,
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B,
		0xFF, 0xC4, 0x00, 0xB5, 0x10, // DHT marker (AC)
		0x00, 0x02, 0x01, 0x03, 0x03, 0x02, 0x04, 0x03,
		0x05, 0x05, 0x04, 0x04, 0x00, 0x00, 0x01, 0x7D,
		0x01, 0x02, 0x03, 0x00, 0x04, 0x11, 0x05, 0x12,
		0x21, 0x31, 0x41, 0x06, 0x13, 0x51, 0x61, 0x07,
		0x22, 0x71, 0x14, 0x32, 0x81, 0x91, 0xA1, 0x08,
		0x23, 0x42, 0xB1, 0xC1, 0x15, 0x52, 0xD1, 0xF0,
		0x24, 0x33, 0x62, 0x72, 0x82, 0x09, 0x0A, 0x16,
		0x17, 0x18, 0x19, 0x1A, 0x25, 0x26, 0x27, 0x28,
		0x29, 0x2A, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39,
		0x3A, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49,
		0x4A, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59,
		0x5A, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69,
		0x6A, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79,
		0x7A, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89,
		0x8A, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98,
		0x99, 0x9A, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7,
		0xA8, 0xA9, 0xAA, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6,
		0xB7, 0xB8, 0xB9, 0xBA, 0xC2, 0xC3, 0xC4, 0xC5,
		0xC6, 0xC7, 0xC8, 0xC9, 0xCA, 0xD2, 0xD3, 0xD4,
		0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xE1, 0xE2,
		0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9, 0xEA,
		0xF1, 0xF2, 0xF3, 0xF4, 0xF5, 0xF6, 0xF7, 0xF8,
		0xF9, 0xFA,
		0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01, 0x00, 0x00, 0x3F, 0x00, // SOS marker
		0x7F, 0xFF, // Scan data
		0xFF, 0xD9, // EOI (End of Image)
	}
}

// createTempFile creates a temporary file with the given pattern and content.
func createTempFile(t *testing.T, pattern string, content []byte) string {
	t.Helper()

	f, err := os.CreateTemp("", pattern)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	if _, err := f.Write(content); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		t.Fatalf("failed to write to temp file: %v", err)
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(f.Name())
		t.Fatalf("failed to close temp file: %v", err)
	}

	return f.Name()
}

// createFileInDir creates a file in the specified directory.
func createFileInDir(t *testing.T, dir, name string, content []byte) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to create file %s: %v", path, err)
	}

	return path
}
