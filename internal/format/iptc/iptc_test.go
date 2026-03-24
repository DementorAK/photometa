package iptc

import (
	"encoding/binary"
	"testing"

	"github.com/DementorAK/photometa/internal/platform/locale"
)

func TestDecode(t *testing.T) {
	author := "John Doe"
	caption := "A nice photo"

	data := make([]byte, 0)

	// Tag 1: By-line (2:80)
	data = append(data, 0x1C, 2, 80)
	size := make([]byte, 2)
	binary.BigEndian.PutUint16(size, uint16(len(author)))
	data = append(data, size...)
	data = append(data, author...)

	// Tag 2: Caption-Abstract (2:120)
	data = append(data, 0x1C, 2, 120)
	size2 := make([]byte, 2)
	binary.BigEndian.PutUint16(size2, uint16(len(caption)))
	data = append(data, size2...)
	data = append(data, caption...)

	// Add garbage bytes before first tag to test skipping
	payload := append([]byte{0x00, 0x01}, data[:13]...)
	payload = append(payload, 0xFF)
	payload = append(payload, data[13:]...)

	tags, err := Decode(payload)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}

	// Verify By-line
	if tags[0].ID() != "2:80" {
		t.Errorf("Tag 0 ID = %s, want 2:80", tags[0].ID())
	}
	if val := tags[0].RawValue(); val != "John Doe" {
		t.Errorf("Tag 0 Value = %v, want John Doe", val)
	}
	if name := tags[0].SpecName(); name != "By-line" {
		t.Errorf("Tag 0 SpecName = %s, want By-line", name)
	}

	// Verify Caption-Abstract
	if tags[1].ID() != "2:120" {
		t.Errorf("Tag 1 ID = %s, want 2:120", tags[1].ID())
	}
}

func TestIPTC_RussianLocale(t *testing.T) {
	locale.SetLocale("ru")
	defer locale.SetLocale("en")

	tag := Tag{record: 2, dataset: 80}
	if name := tag.Name(); name != "Автор" {
		t.Errorf("Name(ru) = %s, want Автор", name)
	}
	if name := tag.SpecName(); name != "By-line" {
		t.Errorf("SpecName = %s, want By-line", name)
	}
}

func TestIPTC_DefaultLocale(t *testing.T) {
	locale.SetLocale("en")
	tag := Tag{record: 2, dataset: 80}
	if name := tag.Name(); name != "By-line" {
		t.Errorf("Name(en) = %s, want By-line", name)
	}
}

func TestDecode_Invalid(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Truncated Header", []byte{0x1C, 2, 80, 0}},
		{"Truncated Data", []byte{0x1C, 2, 80, 0, 5, 'J', 'o', 'h'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, err := Decode(tt.data)
			if err == nil && len(tags) > 0 {
				t.Errorf("expected error or zero tags, got %d tags", len(tags))
			}
		})
	}
}
