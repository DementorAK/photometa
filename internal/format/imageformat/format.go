// Package imageformat provides image format detection from binary data
// using magic bytes and content-type sniffing.
package imageformat

import (
	"bytes"
	"net/http"
	"strings"
)

// Format represents a recognized image file format.
type Format string

const (
	JPEG    Format = "jpeg"
	PNG     Format = "png"
	GIF     Format = "gif"
	WebP    Format = "webp"
	TIFF    Format = "tiff"
	Unknown Format = "unknown"
)

// AllFormats returns all supported image formats.
func AllFormats() []Format {
	return []Format{JPEG, PNG, GIF, WebP, TIFF, Unknown}
}

// Detect determines the image format from file content using magic bytes.
// It first tries [http.DetectContentType], then falls back to manual
// signature checks for formats not always recognized by the standard library
// (WebP, TIFF).
//
// Returns [Unknown] for empty data or unrecognized content.
func Detect(data []byte) Format {
	if len(data) == 0 {
		return Unknown
	}

	contentType := http.DetectContentType(data)

	switch {
	case strings.HasPrefix(contentType, "image/jpeg"):
		return JPEG
	case strings.HasPrefix(contentType, "image/png"):
		return PNG
	case strings.HasPrefix(contentType, "image/gif"):
		return GIF
	case strings.HasPrefix(contentType, "image/webp"):
		return WebP
	case strings.HasPrefix(contentType, "image/tiff"):
		return TIFF
	default:
		return detectManual(data)
	}
}

// detectManual checks signatures not always caught by http.DetectContentType.
func detectManual(data []byte) Format {
	// WebP: RIFF....WEBP
	if len(data) >= 12 {
		if bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
			return WebP
		}
	}

	// TIFF: II*\x00 (little-endian) or MM\x00* (big-endian)
	if len(data) >= 4 {
		if (data[0] == 0x49 && data[1] == 0x49 && data[2] == 0x2A && data[3] == 0x00) ||
			(data[0] == 0x4D && data[1] == 0x4D && data[2] == 0x00 && data[3] == 0x2A) {
			return TIFF
		}
	}

	return Unknown
}
