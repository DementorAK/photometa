package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/DementorAK/photometa/internal/domain"
)

func TestMetadataJSONSerialization(t *testing.T) {
	img := domain.ImageFile{
		Path: "/images/IMG_1234.jpg",
		Name: "IMG_1234.jpg",
		Metadata: domain.Metadata{
			FileSize: 2048576,
			Format:   "jpeg",
			Tags: []domain.TagInfo{
				{Type: "EXIF", Group: "Equipment", Name: "Make", Value: "Canon"},
				{Type: "EXIF", Group: "Equipment", Name: "Model", Value: "EOS R5"},
				{Type: "EXIF", Group: "Shooting", Name: "ISO", Value: 100},
				{Type: "EXIF", Group: "Shooting", Name: "FNumber", Value: 2.8},
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(img)
	if err != nil {
		t.Fatalf("failed to marshal ImageFile: %v", err)
	}

	// Unmarshal back
	var decoded domain.ImageFile
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ImageFile: %v", err)
	}

	// Verify key fields
	if decoded.Name != img.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, img.Name)
	}
	if decoded.Metadata.Format != "jpeg" {
		t.Errorf("Format mismatch: got %q, want %q", decoded.Metadata.Format, "jpeg")
	}

	if len(decoded.Metadata.Tags) != 4 {
		t.Fatalf("Tags length mismatch: got %d, want 4", len(decoded.Metadata.Tags))
	}
}
