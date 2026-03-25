package filler

import (
	"bytes"
	"testing"
	"time"

	"github.com/DementorAK/photometa/internal/domain"
	"github.com/DementorAK/photometa/internal/fake"
)

const (
	tagHeight = "Height"
	tagWidth  = "Width"
)

func TestExtractSegmentsJPEG_Invalid(t *testing.T) {
	logger := fake.NewMockLogger()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "too short",
			data:    []byte{0x00, 0x00},
			wantErr: true,
		},
		{
			name:    "not a JPEG - wrong magic",
			data:    []byte{0x00, 0x01, 0x02},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segs, err := extractSegmentsJPEG(bytes.NewReader(tt.data), logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractSegmentsJPEG() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && (segs.exif != nil || segs.xmp != nil || segs.iptc != nil) {
				t.Error("expected no segments on error")
			}
		})
	}
}

func TestExtractSegmentsJPEG_Valid(t *testing.T) {
	logger := fake.NewMockLogger()

	jpeg := []byte{
		0xFF, 0xD8, // SOI
		0xFF, 0xE0, 0x00, 0x10, // APP0 marker + length
		0x4A, 0x46, 0x49, 0x46, 0x00, // "JFIF\0"
		0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, // JFIF data
		0xFF, 0xC0, // SOF0 marker
		0x00, 0x0B, // Length = 11
		0x08,       // Precision = 8 bits
		0x00, 0x80, // Height = 128
		0x01, 0x00, // Width = 256
		0x01,             // Number of components = 1
		0x01, 0x11, 0x00, // Component 1
		0xFF, 0xD9, // EOI
	}

	segs, err := extractSegmentsJPEG(bytes.NewReader(jpeg), logger)
	if err != nil {
		t.Fatalf("extractSegmentsJPEG() unexpected error: %v", err)
	}

	if segs.width != 256 {
		t.Errorf("expected width 256, got %d", segs.width)
	}
	if segs.height != 128 {
		t.Errorf("expected height 128, got %d", segs.height)
	}
}

func TestExtractSegmentsPNG_Invalid(t *testing.T) {
	logger := fake.NewMockLogger()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "too short",
			data:    []byte{0x00, 0x01},
			wantErr: true,
		},
		{
			name:    "not a PNG - wrong signature",
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

func TestExtractSegmentsPNG_Valid(t *testing.T) {
	logger := fake.NewMockLogger()

	png := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR length
		0x49, 0x48, 0x44, 0x52, // "IHDR"
		0x00, 0x00, 0x01, 0x00, // Width = 256
		0x00, 0x00, 0x00, 0x80, // Height = 128
		0x08,             // Bit depth = 8
		0x02,             // Color type = RGB
		0x00, 0x00, 0x00, // Compression, filter, interlace
		0x90, 0x77, 0x53, 0xDE, // CRC
		0x00, 0x00, 0x00, 0x00, // IEND length
		0x49, 0x45, 0x4E, 0x44, // "IEND"
		0xAE, 0x42, 0x60, 0x82, // CRC
	}

	segs, err := extractSegmentsPNG(bytes.NewReader(png), logger)
	if err != nil {
		t.Fatalf("extractSegmentsPNG() unexpected error: %v", err)
	}

	if segs.width != 256 {
		t.Errorf("expected width 256, got %d", segs.width)
	}
	if segs.height != 128 {
		t.Errorf("expected height 128, got %d", segs.height)
	}
}

func TestExtractSegmentsWebP_Invalid(t *testing.T) {
	logger := fake.NewMockLogger()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "too short",
			data:    []byte{0x00},
			wantErr: true,
		},
		{
			name:    "not a WebP - RIFF without WEBP",
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
		t.Fatalf("extractSegmentsTIFF() unexpected error: %v", err)
	}
	if !bytes.Equal(segs.exif, data) {
		t.Errorf("expected TIFF data in exif field")
	}
}

func TestExtractIPTCFromIRB(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "not 8BIM",
			data: []byte("8XXX"),
		},
		{
			name: "truncated",
			data: []byte("8BIM"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractIPTCFromIRB(tt.data)
			if got != nil {
				t.Errorf("expected nil for %q, got %v", tt.name, got)
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
		{"float32", float32(2.71), 2.7100000381469727},
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

	jpeg := []byte{
		0xFF, 0xD8, // SOI
		0xFF, 0xE0, 0x00, 0x10, // APP0 marker + length
		0x4A, 0x46, 0x49, 0x46, 0x00, // "JFIF\0"
		0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, // thumbnail
		0xFF, 0xC0, // SOF0 marker
		0x00, 0x0B, // Length = 11
		0x08,       // Precision = 8 bits
		0x00, 0x80, // Height = 128
		0x01, 0x00, // Width = 256
		0x01,             // Number of components = 1
		0x01, 0x11, 0x00, // Component 1
		0xFF, 0xD9, // EOI
	}

	reader := bytes.NewReader(jpeg)
	Fill(reader, meta, logger)

	if len(meta.Tags) == 0 {
		t.Fatal("expected Tags to be populated")
	}

	foundH, foundW := false, false
	for _, tag := range meta.Tags {
		if tag.Name == tagHeight {
			foundH = true
		}
		if tag.Name == tagWidth {
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

func TestFill_TIFF_InvalidData(t *testing.T) {
	logger := fake.NewMockLogger()
	meta := &domain.Metadata{
		Format: "tiff",
	}

	tiff := []byte("not valid tiff data")
	reader := bytes.NewReader(tiff)

	Fill(reader, meta, logger)

	if len(logger.ErrorCalls) == 0 {
		t.Error("expected error for invalid TIFF data")
	}
}

func TestFill_TIFF_ValidStructure(t *testing.T) {
	logger := fake.NewMockLogger()
	meta := &domain.Metadata{
		Format: "tiff",
	}

	tiff := createMinimalTIFF()
	reader := bytes.NewReader(tiff)

	Fill(reader, meta, logger)

	if len(logger.ErrorCalls) != 0 {
		t.Errorf("expected no errors for valid TIFF, got %d", len(logger.ErrorCalls))
	}
}

func createMinimalTIFF() []byte {
	tiff := []byte{
		0x49, 0x49, 0x2A, 0x00, // Little-endian TIFF magic
		0x08, 0x00, 0x00, 0x00, // Offset to first IFD
	}

	ifd := []byte{
		0x00, 0x00, // Number of entries = 0
		0x00, 0x00, 0x00, 0x00, // Next IFD offset = 0 (no next IFD)
	}

	return append(tiff, ifd...)
}

func TestFill_PNG_InvalidChunk(t *testing.T) {
	logger := fake.NewMockLogger()
	meta := &domain.Metadata{
		Format: "png",
	}

	png := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x00, // Truncated IHDR length
	}

	reader := bytes.NewReader(png)
	Fill(reader, meta, logger)

	if len(meta.Tags) != 0 {
		t.Error("expected no Tags for invalid PNG")
	}
}

func TestFill_WebP(t *testing.T) {
	logger := fake.NewMockLogger()
	meta := &domain.Metadata{
		Format: "webp",
	}

	webp := []byte{
		0x52, 0x49, 0x46, 0x46, // "RIFF"
		0x1C, 0x00, 0x00, 0x00, // RIFF chunk size (file size - 8)
		0x57, 0x45, 0x42, 0x50, // "WEBP"
		0x56, 0x50, 0x38, 0x58, // "VP8X"
		0x0A, 0x00, 0x00, 0x00, // VP8X chunk size = 10
		0x00,             // Flags (no EXIF, XMP, alpha)
		0x00, 0x00, 0x00, // Reserved
		0xFF, 0x03, 0x00, // Width-1 = 1023 (little-endian 24-bit)
		0xFF, 0x01, 0x00, // Height-1 = 511 (little-endian 24-bit)
		0x56, 0x50, 0x38, 0x20, // "VP8 " (simple lossless)
		0x08, 0x00, 0x00, 0x00, // VP8 chunk size
		0x00, 0x00, 0x00, 0x00, // VP8 data
		0x00, 0x00, 0x00, 0x00, // Padding
	}

	reader := bytes.NewReader(webp)
	Fill(reader, meta, logger)

	if len(meta.Tags) == 0 {
		t.Error("expected Tags to be populated for valid WebP")
	}

	foundH, foundW := false, false
	for _, tag := range meta.Tags {
		if tag.Name == tagHeight {
			foundH = true
		}
		if tag.Name == tagWidth {
			foundW = true
		}
	}

	if !foundH || !foundW {
		t.Error("expected Height and Width Tags for WebP")
	}
}

func TestFill_JPEG_WithEXIF(t *testing.T) {
	logger := fake.NewMockLogger()
	meta := &domain.Metadata{
		Format:   "jpeg",
		FileSize: 2048,
	}

	jpeg := createJPEGWithEXIF()
	reader := bytes.NewReader(jpeg)

	Fill(reader, meta, logger)

	if len(logger.ErrorCalls) != 0 {
		t.Errorf("expected no errors, got %d", len(logger.ErrorCalls))
	}
}

func createJPEGWithEXIF() []byte {
	jpeg := []byte{
		0xFF, 0xD8, // SOI
		0xFF, 0xE0, 0x00, 0x10, // APP0 marker + length
		0x4A, 0x46, 0x49, 0x46, 0x00, // "JFIF\0"
		0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, // JFIF data
	}

	exif := []byte{
		0xFF, 0xE1, // APP1 marker
		0x00, 0x20, // Length = 32
		0x45, 0x78, 0x69, 0x66, 0x00, 0x00, // "Exif\0\0"
		0x49, 0x49, 0x2A, 0x00, // Little-endian TIFF magic
		0x08, 0x00, 0x00, 0x00, // Offset to first IFD
		0x00, 0x00, // IFD with 0 entries
		0x00, 0x00, 0x00, 0x00, // Next IFD = 0
	}

	jpeg = append(jpeg, exif...)

	jpeg = append(jpeg, []byte{
		0xFF, 0xC0, // SOF0 marker
		0x00, 0x0B, // Length = 11
		0x08,       // Precision = 8 bits
		0x00, 0x80, // Height = 128
		0x01, 0x00, // Width = 256
		0x01,             // Number of components = 1
		0x01, 0x11, 0x00, // Component 1
		0xFF, 0xD9, // EOI
	}...)

	return jpeg
}

func TestFill_EmptyReader(t *testing.T) {
	logger := fake.NewMockLogger()
	meta := &domain.Metadata{
		Format: "jpeg",
	}
	reader := bytes.NewReader([]byte{})

	Fill(reader, meta, logger)

	if len(meta.Tags) != 0 {
		t.Error("expected no Tags for empty reader")
	}
}

func TestFill_JPEG_SkipUnsupportedMarkers(t *testing.T) {
	logger := fake.NewMockLogger()
	meta := &domain.Metadata{
		Format:   "jpeg",
		FileSize: 1024,
	}

	jpeg := []byte{
		0xFF, 0xD8, // SOI
		0xFF, 0xE0, 0x00, 0x10, // APP0 marker + length
		0x4A, 0x46, 0x49, 0x46, 0x00, // "JFIF\0"
		0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, // thumbnail
		0xFF, 0xDB, // DQT marker (unsupported)
		0x00, 0x43, // Length = 67
	}

	for i := 0; i < 65; i++ {
		jpeg = append(jpeg, 0x00)
	}

	jpeg = append(jpeg, []byte{
		0xFF, 0xC0, // SOF0 marker
		0x00, 0x0B, // Length = 11
		0x08,       // Precision = 8 bits
		0x00, 0x80, // Height = 128
		0x01, 0x00, // Width = 256
		0x01,             // Number of components = 1
		0x01, 0x11, 0x00, // Component 1
		0xFF, 0xD9, // EOI
	}...)

	reader := bytes.NewReader(jpeg)
	Fill(reader, meta, logger)

	foundH, foundW := false, false
	for _, tag := range meta.Tags {
		if tag.Name == tagHeight {
			foundH = true
		}
		if tag.Name == tagWidth {
			foundW = true
		}
	}

	if !foundH || !foundW {
		t.Error("expected Height and Width Tags despite unsupported markers")
	}
}
