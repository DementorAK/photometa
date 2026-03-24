package imageformat

import "testing"

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected Format
	}{
		{
			name: "JPEG",
			data: []byte{
				0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10,
				0x4A, 0x46, 0x49, 0x46, 0x00,
			},
			expected: JPEG,
		},
		{
			name:     "PNG",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x00},
			expected: PNG,
		},
		{
			name:     "GIF",
			data:     []byte("GIF89a"),
			expected: GIF,
		},
		{
			name:     "WebP",
			data:     []byte("RIFF\x00\x00\x00\x00WEBP"),
			expected: WebP,
		},
		{
			name:     "TIFF_LittleEndian",
			data:     []byte{0x49, 0x49, 0x2A, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: TIFF,
		},
		{
			name:     "TIFF_BigEndian",
			data:     []byte{0x4D, 0x4D, 0x00, 0x2A, 0x00, 0x00, 0x00, 0x00},
			expected: TIFF,
		},
		{
			name:     "Unknown",
			data:     []byte("random data here"),
			expected: Unknown,
		},
		{
			name:     "Empty",
			data:     []byte{},
			expected: Unknown,
		},
		{
			name:     "Nil",
			data:     nil,
			expected: Unknown,
		},
		{
			name:     "TooShort",
			data:     []byte{0x00, 0x01},
			expected: Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Detect(tt.data)
			if result != tt.expected {
				t.Errorf("Detect(%s) = %q, want %q", tt.name, result, tt.expected)
			}
		})
	}
}

func TestFormatStringValues(t *testing.T) {
	// Ensure format constants have expected string values.
	if string(JPEG) != "jpeg" {
		t.Errorf("JPEG = %q, want %q", JPEG, "jpeg")
	}
	if string(Unknown) != "unknown" {
		t.Errorf("Unknown = %q, want %q", Unknown, "unknown")
	}
}
