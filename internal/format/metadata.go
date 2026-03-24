// Package metadata defines the universal contracts for image metadata tags.
//
// All metadata sources (EXIF, IPTC, XMP) implement the [Tag] interface,
// allowing uniform access to tag identity, localized names, and values.
//
// Locale management is provided by the [github.com/DementorAK/photometa/internal/platform/locale] package.
package format

import "fmt"

// Source identifies the origin of a metadata tag.
type Source int

const (
	SourceEXIF Source = iota
	SourceIPTC
	SourceXMP
)

// String returns the human-readable name of the source.
func (s Source) String() string {
	switch s {
	case SourceEXIF:
		return "EXIF"
	case SourceIPTC:
		return "IPTC"
	case SourceXMP:
		return "XMP"
	default:
		return fmt.Sprintf("Unknown(%d)", int(s))
	}
}

// Tag is the universal contract for any metadata tag,
// regardless of its source (EXIF, IPTC, XMP, or future formats).
type Tag interface {
	// Source returns the metadata source (EXIF, IPTC, XMP).
	Source() Source

	// ID returns a unique identifier for the tag within its source.
	// For EXIF: hex tag ID like "0x010F".
	// For IPTC: "record:dataset" like "2:80".
	// For XMP: namespace-qualified name like "dc:creator".
	ID() string

	// Name returns the human-readable tag name in the current
	// global locale (see [SetLocale]). If no translation is available,
	// returns the specification name (English).
	Name() string

	// SpecName returns the canonical English name from the specification.
	// This value never changes regardless of the current locale.
	SpecName() string

	// RawValue returns the tag value in its native Go type.
	RawValue() any

	// String returns the tag value formatted as a display string.
	String() string
}
