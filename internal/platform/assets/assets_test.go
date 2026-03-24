package assets

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DementorAK/photometa/internal/format/imageformat"
	"github.com/DementorAK/photometa/internal/domain"
)

func TestGetIcon(t *testing.T) {
	// Test one format and one group to verify mechanics
	iconNames := []string{"format_png", "group_author"}

	for _, name := range iconNames {
		t.Run(name, func(t *testing.T) {
			data, err := GetIcon(name)
			if err != nil {
				t.Fatalf("GetIcon(%s) error = %v", name, err)
			}
			if len(data) == 0 {
				t.Error("Empty data returned")
			}
			if !strings.Contains(string(data), "<svg") {
				t.Error("Invalid SVG content")
			}
		})
	}
}

func TestIconConsistency(t *testing.T) {
	// Verify icons for all supported image formats
	for _, format := range imageformat.AllFormats() {
		name := fmt.Sprintf("format_%s", format)
		t.Run(name, func(t *testing.T) {
			data, err := GetIcon(name)
			if err != nil {
				t.Errorf("Missing icon for format: %s", format)
				return
			}
			validateSVG(t, name, data)
		})
	}

	// Verify icons for all metadata property groups
	for _, group := range domain.AllGroups() {
		name := fmt.Sprintf("group_%s", strings.ToLower(group))
		t.Run(name, func(t *testing.T) {
			data, err := GetIcon(name)
			if err != nil {
				t.Errorf("Missing icon for group: %s", group)
				return
			}
			validateSVG(t, name, data)
		})
	}
}

func validateSVG(t *testing.T, name string, data []byte) {
	t.Helper()
	s := string(data)
	if !strings.Contains(s, "<svg") || !strings.Contains(s, "</svg>") {
		t.Errorf("Icon %s has invalid SVG content", name)
	}
	if !strings.Contains(s, "<?xml") {
		t.Errorf("Icon %s is missing XML header", name)
	}
}

func TestGetIcon_Error(t *testing.T) {
	_, err := GetIcon("non_existent_icon")
	if err == nil {
		t.Error("GetIcon(non_existent_icon) expected error, got nil")
	}
}

func TestListIcons(t *testing.T) {
	icons, err := ListIcons()
	if err != nil {
		t.Fatalf("ListIcons() error = %v", err)
	}

	// Total expected icons: formats + groups
	expectedCount := len(imageformat.AllFormats()) + len(domain.AllGroups())
	if len(icons) != expectedCount {
		t.Errorf("ListIcons() returned %d icons, expected %d", len(icons), expectedCount)
	}
}
