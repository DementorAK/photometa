package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/DementorAK/photometa/internal/analyzer"
	"github.com/DementorAK/photometa/internal/fake"
)

func TestIntegration_AnalyzeWebP(t *testing.T) {
	logger := fake.NewMockLogger()
	service := analyzer.NewService(logger)

	cwd, _ := os.Getwd()
	webpPath := filepath.Join(cwd, "..", "docs", "img", "sample_webp.webp")
	if _, err := os.Stat(webpPath); os.IsNotExist(err) {
		webpPath = filepath.Join(cwd, "docs", "img", "sample_webp.webp")
	}

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

	// sample_webp.webp is known to have some tags (from previous conversations context)
	if len(result.Metadata.Tags) == 0 {
		t.Error("Expected metadata tags for sample_webp.webp, got none")
	}

	t.Logf("Successfully analyzed WebP. Tags count: %d", len(result.Metadata.Tags))
}

func TestIntegration_AnalyzeDirectory(t *testing.T) {
	logger := fake.NewMockLogger()
	service := analyzer.NewService(logger)

	cwd, _ := os.Getwd()
	imgDir := filepath.Join(cwd, "..", "docs", "img")
	if _, err := os.Stat(imgDir); os.IsNotExist(err) {
		imgDir = filepath.Join(cwd, "docs", "img")
	}

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

	var foundLogo, foundREADME bool
	for _, res := range results {
		if res.Name == "README.md" {
			foundREADME = true
		}
		if res.Name == "photometa_logo.jpg" {
			foundLogo = true
			// Check that it has NO meta-tags (EXIF, IPTC, XMP)
			for _, tag := range res.Metadata.Tags {
				if tag.Type == "EXIF" || tag.Type == "IPTC" || tag.Type == "XMP" {
					t.Errorf("Found unexpected meta-tag %q of type %q in photometa_logo.jpg", tag.Name, tag.Type)
				}
			}
		}
	}

	if foundREADME {
		t.Error("Scan results should NOT contain README.md")
	}

	if !foundLogo {
		t.Error("Scan results should contain photometa_logo.jpg")
	}

	t.Logf("Directory scan successful. Found %d images.", len(results))
}
