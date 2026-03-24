package filler

import (
	"bytes"
	"testing"
	"time"

	"github.com/DementorAK/photometa/internal/domain"
	"github.com/DementorAK/photometa/internal/fake"
)

func TestExtractSegmentsJPEG(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantErr  bool
		wantEXIF bool
		wantXMP  bool
		wantSize int
	}{
		{
			name:    "invalid data",
			data:    []byte{0x00, 0x00},
			wantErr: true,
		},
		{
			name:    "not a JPEG",
			data:    []byte{0x00, 0x01, 0x02},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segs, err := extractSegmentsJPEG(bytes.NewReader(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("extractSegmentsJPEG() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantEXIF && len(segs.exif) == 0 {
				t.Error("expected EXIF data")
			}
			if tt.wantXMP && len(segs.xmp) == 0 {
				t.Error("expected XMP data")
			}
		})
	}
}

func TestExtractSegmentsPNG(t *testing.T) {
	logger := fake.NewMockLogger()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "invalid data",
			data:    []byte{0x00, 0x01},
			wantErr: true,
		},
		{
			name:    "not a PNG",
			data:    []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractSegmentsPNG(bytes.NewReader(tt.data), logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractSegmentsPNG() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractSegmentsWebP(t *testing.T) {
	logger := fake.NewMockLogger()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "invalid data",
			data:    []byte{0x00},
			wantErr: true,
		},
		{
			name:    "not a WebP",
			data:    []byte("RIFF"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractSegmentsWebP(bytes.NewReader(tt.data), logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractSegmentsWebP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractSegmentsTIFF(t *testing.T) {
	data := []byte("TIFF data")
	segs, err := extractSegmentsTIFF(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(segs.exif, data) {
		t.Errorf("expected TIFF data in exif field")
	}
}

func TestExtractIPTCFromIRB(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []byte
	}{
		{
			name: "empty data",
			data: []byte{},
			want: nil,
		},
		{
			name: "not 8BIM",
			data: []byte("8XXX"),
			want: nil,
		},
		{
			name: "truncated",
			data: []byte("8BIM"),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractIPTCFromIRB(tt.data)
			if tt.want == nil && got != nil {
				t.Errorf("expected nil, got %v", got)
			}
		})
	}
}

func TestTagToFloat(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  float64
	}{
		{"float64", float64(3.14), 3.14},
		{"float32", float32(2.71), 2.7100000381469727}, // float32 to float64 conversion
		{"int", int(42), 42.0},
		{"int16", int16(10), 10.0},
		{"int32", int32(20), 20.0},
		{"uint16", uint16(30), 30.0},
		{"uint32", uint32(40), 40.0},
		{"[]float64", []float64{1.5, 2.5}, 1.5},
		{"[]uint32", []uint32{100, 200}, 100.0},
		{"[]float64 empty", []float64{}, 0.0},
		{"unknown", "string", 0.0},
		{"nil", nil, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TagToFloat(tt.input)
			if got != tt.want {
				t.Errorf("TagToFloat(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTagToInt(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  int
	}{
		{"int", int(42), 42},
		{"int64", int64(100), 100},
		{"float64", float64(99.9), 99},
		{"int16", int16(10), 10},
		{"int32", int32(20), 20},
		{"uint16", uint16(30), 30},
		{"uint32", uint32(40), 40},
		{"[]uint16", []uint16{50, 60}, 50},
		{"[]uint16 empty", []uint16{}, 0},
		{"unknown", "string", 0},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TagToInt(tt.input)
			if got != tt.want {
				t.Errorf("TagToInt(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTagToTime(t *testing.T) {
	exifTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name string
		val  any
		str  string
		want *time.Time
	}{
		{"time.Time", exifTime, "", &exifTime},
		{"EXIF format", nil, "2024:01:15 10:30:00", &exifTime},
		{"XMP format", nil, "2024-01-15T10:30:00", &exifTime},
		{"null terminated", nil, "2024:01:15 10:30:00\x00\x00", &exifTime},
		{"too short", nil, "2024:01", nil},
		{"invalid format", nil, "15-01-2024 10:30:00", nil},
		{"nil values", nil, "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TagToTime(tt.val, tt.str)
			if tt.want == nil {
				if got != nil {
					t.Errorf("TagToTime() = %v, want nil", *got)
				}
			} else {
				if got == nil {
					t.Errorf("TagToTime() = nil, want %v", *tt.want)
				} else if !got.Equal(*tt.want) {
					t.Errorf("TagToTime() = %v, want %v", *got, *tt.want)
				}
			}
		})
	}
}


func TestFill_UnknownFormat(t *testing.T) {
	logger := fake.NewMockLogger()
	meta := &domain.Metadata{
		Format: "unknown",
	}
	reader := bytes.NewReader([]byte{0x00, 0x01})

	Fill(reader, meta, logger)

	if len(logger.WarnCalls) == 0 {
		t.Error("expected warning for unknown format")
	}
}

func TestFill_ValidJPEG(t *testing.T) {
	logger := fake.NewMockLogger()
	meta := &domain.Metadata{
		Format:   "jpeg",
		FileSize: 1024,
	}

	// Valid minimal JPEG with SOF0 marker containing dimensions
	// SOF0 marker format: FF C0 [length hi lo] [precision] [height hi lo] [width hi lo] ...
	jpeg := []byte{
		0xFF, 0xD8, // SOI
		0xFF, 0xE0, 0x00, 0x10, // APP0 marker + length
		0x4A, 0x46, 0x49, 0x46, 0x00, // "JFIF\0"
		0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, // thumbnail
		0xFF, 0xC0, // SOF0 marker
		0x00, 0x0B, // Length = 11
		0x08,       // Precision = 8 bits
		0x00, 0x80, // Height = 128 (0x0080 = 128)
		0x01, 0x00, // Width = 256 (0x0100 = 256)
		0x01,             // Number of components = 1
		0x01, 0x11, 0x00, // Component 1
		0xFF, 0xD9, // EOI
	}

	reader := bytes.NewReader(jpeg)
	Fill(reader, meta, logger)

	if len(meta.Tags) == 0 {
		t.Fatal("expected Tags to be populated")
	}
	
	// Default JPEG detector adds dimensions as format tags
	foundH, foundW := false, false
	for _, tag := range meta.Tags {
		if tag.Name == "Height" {
			foundH = true
		}
		if tag.Name == "Width" {
			foundW = true
		}
	}
	
	if !foundH {
		t.Errorf("expected Height Tag")
	}
	if !foundW {
		t.Errorf("expected Width Tag")
	}
}

func TestFill_TIFF(t *testing.T) {
	logger := fake.NewMockLogger()
	meta := &domain.Metadata{
		Format: "tiff",
	}

	tiff := []byte("some tiff data")
	reader := bytes.NewReader(tiff)

	Fill(reader, meta, logger)

	if logger.ErrorCalls != nil {
		t.Error("Fill() should not log errors for TIFF")
	}
}
