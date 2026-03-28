package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/DementorAK/photometa/internal/analyzer"
	"github.com/DementorAK/photometa/internal/domain"
	"github.com/DementorAK/photometa/internal/fake"
)

func hasTagType(tags []domain.TagInfo, tagType string) bool {
	for _, tag := range tags {
		if tag.Type == tagType {
			return true
		}
	}
	return false
}

func hasAnyMetadata(tags []domain.TagInfo) bool {
	return hasTagType(tags, "EXIF") || hasTagType(tags, "IPTC") || hasTagType(tags, "XMP")
}

func TestIntegration_AnalyzeDirectory(t *testing.T) {
	logger := fake.NewMockLogger()
	service := analyzer.NewService(logger)

	cwd, _ := os.Getwd()
	imgDir := filepath.Join(cwd, "testdata")

	if _, err := os.Stat(imgDir); os.IsNotExist(err) {
		t.Skipf("Skipping integration test: image directory not found at %s", imgDir)
	}

	ctx := context.Background()
	results, err := service.ScanDirectory(ctx, imgDir)
	if err != nil {
		t.Fatalf("Failed to scan directory: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected at least one image in directory scan, got zero")
	}

	var foundREADME bool
	for _, res := range results {
		if res.Name == "README.md" {
			foundREADME = true
		}
	}

	if foundREADME {
		t.Error("Scan results should NOT contain README.md")
	}

	t.Logf("Directory scan successful. Found %d images.", len(results))
}

func TestIntegration_AllSamplesHaveMetadata(t *testing.T) {
	logger := fake.NewMockLogger()
	service := analyzer.NewService(logger)

	cwd, _ := os.Getwd()
	imgDir := filepath.Join(cwd, "testdata")

	if _, err := os.Stat(imgDir); os.IsNotExist(err) {
		t.Skipf("Skipping integration test: image directory not found at %s", imgDir)
	}

	ctx := context.Background()
	results, err := service.ScanDirectory(ctx, imgDir)
	if err != nil {
		t.Fatalf("Failed to scan directory: %v", err)
	}

	excludedFiles := map[string]bool{
		"README.md": true,
	}

	for _, res := range results {
		if excludedFiles[res.Name] {
			continue
		}

		if !hasAnyMetadata(res.Metadata.Tags) {
			t.Errorf("Expected at least one metadata type (EXIF, IPTC, or XMP) for %s, got none", res.Name)
		}
	}
}

func TestIntegration_SampleJPG1HasAllMetadataTypes(t *testing.T) {
	logger := fake.NewMockLogger()
	service := analyzer.NewService(logger)

	cwd, _ := os.Getwd()
	jpegPath := filepath.Join(cwd, "testdata", "sample_jpg_1.jpg")

	if _, err := os.Stat(jpegPath); os.IsNotExist(err) {
		t.Skipf("Skipping integration test: sample_jpg_1.jpg not found at %s", jpegPath)
	}

	ctx := context.Background()
	result, err := service.AnalyzeFile(ctx, jpegPath)
	if err != nil {
		t.Fatalf("Failed to analyze sample_jpg_1.jpg: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Metadata.Format != "jpeg" {
		t.Errorf("Expected format 'jpeg', got %q", result.Metadata.Format)
	}

	if !hasTagType(result.Metadata.Tags, "EXIF") {
		t.Error("sample_jpg_1.jpg must have EXIF metadata")
	}
	if !hasTagType(result.Metadata.Tags, "IPTC") {
		t.Error("sample_jpg_1.jpg must have IPTC metadata")
	}
	if !hasTagType(result.Metadata.Tags, "XMP") {
		t.Error("sample_jpg_1.jpg must have XMP metadata")
	}

	t.Logf("sample_jpg_1.jpg has all metadata types. Tags count: %d", len(result.Metadata.Tags))
}

func TestIntegration_AnalyzeWebP(t *testing.T) {
	logger := fake.NewMockLogger()
	service := analyzer.NewService(logger)

	cwd, _ := os.Getwd()
	webpPath := filepath.Join(cwd, "testdata", "sample_webp.webp")

	if _, err := os.Stat(webpPath); os.IsNotExist(err) {
		t.Skipf("Skipping integration test: webp image not found at %s", webpPath)
	}

	ctx := context.Background()
	result, err := service.AnalyzeFile(ctx, webpPath)
	if err != nil {
		t.Fatalf("Failed to analyze webp image: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Metadata.Format != "webp" {
		t.Errorf("Expected format 'webp', got %q", result.Metadata.Format)
	}

	if len(result.Metadata.Tags) == 0 {
		t.Error("Expected metadata tags for sample_webp.webp, got none")
	}

	t.Logf("Successfully analyzed WebP. Tags count: %d", len(result.Metadata.Tags))
}
