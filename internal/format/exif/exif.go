package exif

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/DementorAK/photometa/internal/format"
)

// Tag represents a single EXIF metadata tag.
type Tag struct {
	id    uint16
	isGPS bool // true if this tag comes from GPS IFD
	value any
}

// Ensure Tag implements format.Tag interface.
var _ format.Tag = Tag{}

func (t Tag) Source() format.Source {
	return format.SourceEXIF
}

func (t Tag) ID() string {
	return fmt.Sprintf("0x%04X", t.id)
}

// Name returns the localized tag name using the current global locale.
// Falls back to the specification name if no translation is available.
func (t Tag) Name() string {
	return getLocalizedName(t.SpecName())
}

// SpecName returns the canonical English name from the specification.
func (t Tag) SpecName() string {
	return getSpecName(t.id, t.isGPS)
}

func (t Tag) RawValue() any {
	return t.value
}

func (t Tag) String() string {
	return fmt.Sprintf("%v", t.value)
}

// Decode parses EXIF data from a byte slice.
// The data should be the payload of the APP1 segment (starting with "Exif\x00\x00")
// or a raw TIFF header.
func Decode(data []byte) ([]Tag, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("exif data too short")
	}

	// Skip potential "Exif\0\0" header
	offset := 0
	if len(data) >= 6 && string(data[0:6]) == "Exif\x00\x00" {
		offset = 6
	}

	tiffData := data[offset:]
	if len(tiffData) < 8 {
		return nil, fmt.Errorf("invalid TIFF header")
	}

	// Determine byte order
	var byteOrder binary.ByteOrder
	switch {
	case tiffData[0] == 'I' && tiffData[1] == 'I':
		byteOrder = binary.LittleEndian
	case tiffData[0] == 'M' && tiffData[1] == 'M':
		byteOrder = binary.BigEndian
	default:
		return nil, fmt.Errorf("invalid byte order marker: %x%x", tiffData[0], tiffData[1])
	}

	// Check magic number (42)
	if magic := byteOrder.Uint16(tiffData[2:4]); magic != 42 {
		return nil, fmt.Errorf("invalid TIFF magic number: %d", magic)
	}

	// Offset to the first IFD
	firstIFDOffset := byteOrder.Uint32(tiffData[4:8])
	if firstIFDOffset < 8 || int(firstIFDOffset) >= len(tiffData) {
		return nil, fmt.Errorf("invalid first IFD offset: %d", firstIFDOffset)
	}

	var tags []Tag
	visited := make(map[uint32]bool) // Detect loops in IFD chains

	// Process IFD0
	tags, err := processIFD(tiffData, firstIFDOffset, byteOrder, visited, tags, false)
	if err != nil {
		return tags, err // Return partial results
	}

	return tags, nil
}

func processIFD(data []byte, offset uint32, bo binary.ByteOrder, visited map[uint32]bool, tags []Tag, isGPS bool) ([]Tag, error) {
	if visited[offset] {
		return tags, nil
	}
	visited[offset] = true

	if int(offset)+2 > len(data) {
		return tags, fmt.Errorf("IFD offset out of bounds")
	}

	numEntries := bo.Uint16(data[offset : offset+2])
	current := int(offset) + 2

	for i := 0; i < int(numEntries); i++ {
		if current+12 > len(data) {
			return tags, fmt.Errorf("IFD entry out of bounds")
		}

		tagID := bo.Uint16(data[current : current+2])
		dataType := bo.Uint16(data[current+2 : current+4])
		count := bo.Uint32(data[current+4 : current+8])

		// The value or offset to the value
		valueOffsetBytes := data[current+8 : current+12]

		val, err := parseValue(data, dataType, count, valueOffsetBytes, bo)
		if err == nil {
			tags = append(tags, Tag{id: tagID, isGPS: isGPS, value: val})

			// Handle sub-IFDs
			if count == 1 && dataType == TypeLong {
				subOffset, ok := val.(uint32)
				if ok && subOffset > 0 {
					switch tagID {
					case 0x8769: // ExifIFD
						subTags, _ := processIFD(data, subOffset, bo, visited, nil, false)
						tags = append(tags, subTags...)
					case 0x8825: // GPSInfo
						subTags, _ := processIFD(data, subOffset, bo, visited, nil, true)
						tags = append(tags, subTags...)
					}
				}
			}
		}

		current += 12
	}

	return tags, nil
}

// TIFF Data Types
const (
	TypeByte      = 1
	TypeAscii     = 2
	TypeShort     = 3
	TypeLong      = 4
	TypeRational  = 5
	TypeSByte     = 6
	TypeUndefined = 7
	TypeSShort    = 8
	TypeSLong     = 9
	TypeSRational = 10
	TypeFloat     = 11
	TypeDouble    = 12
)

func parseValue(data []byte, typ uint16, count uint32, valOrOffset []byte, bo binary.ByteOrder) (any, error) {
	size := typeSize(typ)
	if size == 0 {
		return nil, fmt.Errorf("unknown type: %d", typ)
	}

	totalSize := int(size * count)

	var rawData []byte
	if totalSize <= 4 {
		rawData = valOrOffset[:totalSize]
	} else {
		offset := bo.Uint32(valOrOffset)
		if int(offset)+totalSize > len(data) {
			return nil, fmt.Errorf("value offset out of bounds")
		}
		rawData = data[offset : int(offset)+totalSize]
	}

	switch typ {
	case TypeAscii:
		s := string(rawData)
		return strings.TrimRight(s, "\x00"), nil
	case TypeShort:
		if count == 1 {
			return bo.Uint16(rawData), nil
		}
		vals := make([]uint16, count)
		for i := 0; i < int(count); i++ {
			vals[i] = bo.Uint16(rawData[i*2:])
		}
		return vals, nil
	case TypeLong:
		if count == 1 {
			return bo.Uint32(rawData), nil
		}
		vals := make([]uint32, count)
		for i := 0; i < int(count); i++ {
			vals[i] = bo.Uint32(rawData[i*4:])
		}
		return vals, nil
	case TypeRational:
		if count == 1 {
			return readRational(rawData, bo), nil
		}
		vals := make([]float64, count)
		for i := 0; i < int(count); i++ {
			vals[i] = readRational(rawData[i*8:], bo)
		}
		return vals, nil
	case TypeSShort:
		if count == 1 {
			return int16(bo.Uint16(rawData)), nil
		}
		vals := make([]int16, count)
		for i := 0; i < int(count); i++ {
			vals[i] = int16(bo.Uint16(rawData[i*2:]))
		}
		return vals, nil
	case TypeSLong:
		if count == 1 {
			return int32(bo.Uint32(rawData)), nil
		}
		vals := make([]int32, count)
		for i := 0; i < int(count); i++ {
			vals[i] = int32(bo.Uint32(rawData[i*4:]))
		}
		return vals, nil
	case TypeSRational:
		if count == 1 {
			return readSRational(rawData, bo), nil
		}
		vals := make([]float64, count)
		for i := 0; i < int(count); i++ {
			vals[i] = readSRational(rawData[i*8:], bo)
		}
		return vals, nil
	case TypeByte, TypeUndefined:
		if count == 1 {
			return rawData[0], nil
		}
		valCopy := make([]byte, len(rawData))
		copy(valCopy, rawData)
		return valCopy, nil
	default:
		return nil, fmt.Errorf("unsupported type: %d", typ)
	}
}

func readRational(b []byte, bo binary.ByteOrder) float64 {
	num := bo.Uint32(b[:4])
	den := bo.Uint32(b[4:])
	if den == 0 {
		return 0
	}
	return float64(num) / float64(den)
}

func readSRational(b []byte, bo binary.ByteOrder) float64 {
	num := int32(bo.Uint32(b[:4]))
	den := int32(bo.Uint32(b[4:]))
	if den == 0 {
		return 0
	}
	return float64(num) / float64(den)
}

func typeSize(typ uint16) uint32 {
	switch typ {
	case TypeByte, TypeAscii, TypeSByte, TypeUndefined:
		return 1
	case TypeShort, TypeSShort:
		return 2
	case TypeLong, TypeSLong, TypeFloat:
		return 4
	case TypeRational, TypeSRational, TypeDouble:
		return 8
	default:
		return 0
	}
}
