package exif

import (
	"encoding/binary"
	"testing"

	"github.com/DementorAK/photometa/internal/platform/locale"
)

const specMake = "Make"

func TestDecode(t *testing.T) {
	// Construct a minimal valid TIFF payload (Little Endian)
	header := []byte{'I', 'I', 42, 0, 8, 0, 0, 0}

	// IFD0 with 3 entries:
	// Entry 1: Make (0x010F), ASCII, Count=5, Offset -> "ACME\0"
	// Entry 2: Model (0x0110), ASCII, Count=6, Offset -> "RX100\0"
	// Entry 3: Orientation (0x0112), SHORT, Count=1, Value=1 (embedded)

	valueData := []byte("ACME\x00RX100\x00")
	makeOffset := uint32(50) // 8 (header) + 2 (count) + 12*3 (entries) + 4 (next) = 50
	modelOffset := uint32(55)

	ifd := make([]byte, 2+12*3+4)
	binary.LittleEndian.PutUint16(ifd[0:], 3)

	// Entry 1: Make
	binary.LittleEndian.PutUint16(ifd[2:], 0x010F)
	binary.LittleEndian.PutUint16(ifd[4:], 2) // ASCII
	binary.LittleEndian.PutUint32(ifd[6:], 5)
	binary.LittleEndian.PutUint32(ifd[10:], makeOffset)

	// Entry 2: Model (Count 6 > 4, uses offset)
	binary.LittleEndian.PutUint16(ifd[14:], 0x0110)
	binary.LittleEndian.PutUint16(ifd[16:], 2)
	binary.LittleEndian.PutUint32(ifd[18:], 6)
	binary.LittleEndian.PutUint32(ifd[22:], modelOffset)

	// Entry 3: Orientation (embedded)
	binary.LittleEndian.PutUint16(ifd[26:], 0x0112)
	binary.LittleEndian.PutUint16(ifd[28:], 3) // SHORT
	binary.LittleEndian.PutUint32(ifd[30:], 1)
	binary.LittleEndian.PutUint16(ifd[34:], 1)

	binary.LittleEndian.PutUint32(ifd[38:], 0) // Next IFD = 0

	data := append(header, ifd...)
	data = append(data, valueData...)

	tags, err := Decode(data, nil)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tags))
	}

	// Verify Make
	if tags[0].ID() != "0x010F" {
		t.Errorf("Tag 0 ID = %s, want 0x010F", tags[0].ID())
	}
	if val, ok := tags[0].RawValue().(string); !ok || val != "ACME" {
		t.Errorf("Tag 0 Value = %v, want ACME", tags[0].RawValue())
	}
	if name := tags[0].SpecName(); name != specMake {
		t.Errorf("Tag 0 SpecName = %s, want Make", name)
	}

	// Verify Model
	if val, ok := tags[1].RawValue().(string); !ok || val != "RX100" {
		t.Errorf("Tag 1 Value = %v, want RX100", tags[1].RawValue())
	}
	if name := tags[1].SpecName(); name != "Model" {
		t.Errorf("Tag 1 SpecName = %s, want Model", name)
	}

	// Verify Orientation
	if val, ok := tags[2].RawValue().(uint16); !ok || val != 1 {
		t.Errorf("Tag 2 Value = %v, want 1", tags[2].RawValue())
	}
}

func TestName_DefaultLocale(t *testing.T) {
	// Default locale is "en", Name() should return spec name
	locale.SetLocale("en")
	tag := Tag{id: 0x010F}
	if name := tag.Name(); name != specMake {
		t.Errorf("Name() with en locale = %s, want Make", name)
	}
}

func TestName_UkrainianLocale(t *testing.T) {
	locale.SetLocale("ua")
	defer locale.SetLocale("en") // restore

	tag := Tag{id: 0x010F}
	if name := tag.Name(); name != "Виробник" {
		t.Errorf("Name() with ua locale = %s, want Виробник", name)
	}

	// SpecName should always be English regardless of locale
	if name := tag.SpecName(); name != specMake {
		t.Errorf("SpecName() = %s, want Make", name)
	}
}

func TestName_MissingTranslation(t *testing.T) {
	locale.SetLocale("ua")
	defer locale.SetLocale("en")

	// ProcessingSoftware (0x000B) — unlikely to have Ukrainian translation
	tag := Tag{id: 0x000B}
	specName := tag.SpecName()
	name := tag.Name()
	// If no translation, Name() should fall back to spec name
	if name != specName {
		t.Errorf("Name() = %s, expected fallback to SpecName = %s", name, specName)
	}
}

func TestName_UnsupportedLocale(t *testing.T) {
	locale.SetLocale("zh")
	defer locale.SetLocale("en")

	tag := Tag{id: 0x010F}
	// No zh.json exists, should fall back to spec name
	if name := tag.Name(); name != specMake {
		t.Errorf("Name() with zh locale = %s, want Make", name)
	}
}

func TestGPSTag(t *testing.T) {
	tag := Tag{id: 0x0002, isGPS: true}
	if name := tag.SpecName(); name != "GPSLatitude" {
		t.Errorf("GPS SpecName = %s, want GPSLatitude", name)
	}

	locale.SetLocale("ua")
	defer locale.SetLocale("en")
	if name := tag.Name(); name != "Широта" {
		t.Errorf("GPS Name(ua) = %s, want Широта", name)
	}
}

func TestDecode_Rational(t *testing.T) {
	header := []byte{'I', 'I', 42, 0, 8, 0, 0, 0}

	valData := make([]byte, 8)
	binary.LittleEndian.PutUint32(valData[0:], 28)
	binary.LittleEndian.PutUint32(valData[4:], 10)

	offset := uint32(8 + 2 + 12 + 4)

	ifd := make([]byte, 2+12+4)
	binary.LittleEndian.PutUint16(ifd[0:], 1)
	binary.LittleEndian.PutUint16(ifd[2:], 0x829D) // FNumber
	binary.LittleEndian.PutUint16(ifd[4:], 5)      // RATIONAL
	binary.LittleEndian.PutUint32(ifd[6:], 1)
	binary.LittleEndian.PutUint32(ifd[10:], offset)

	data := append(header, ifd...)
	data = append(data, valData...)

	tags, err := Decode(data, nil)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(tags))
	}

	if val, ok := tags[0].RawValue().(float64); !ok || val != 2.8 {
		t.Errorf("expected 2.8, got %v", tags[0].RawValue())
	}
	if name := tags[0].SpecName(); name != "FNumber" {
		t.Errorf("SpecName = %s, want FNumber", name)
	}
}

func TestDecode_WithExifHeader(t *testing.T) {
	prefix := []byte{'E', 'x', 'i', 'f', 0, 0}
	header := []byte{'I', 'I', 42, 0, 8, 0, 0, 0}
	ifd := []byte{0, 0, 0, 0, 0, 0}

	data := append(prefix, header...)
	data = append(data, ifd...)

	tags, err := Decode(data, nil)
	if err != nil {
		t.Fatalf("Decode failed with prefix: %v", err)
	}

	if len(tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(tags))
	}
}

func TestDecode_Invalid(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Empty", []byte{}},
		{"Short", []byte{'I', 'I', 42}},
		{"BadMagic", []byte{'I', 'I', 43, 0, 8, 0, 0, 0}},
		{"BadOrder", []byte{'X', 'Y', 42, 0, 8, 0, 0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decode(tt.data, nil)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}
