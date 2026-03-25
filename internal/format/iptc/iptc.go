package iptc

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/DementorAK/photometa/internal/format"
	"github.com/DementorAK/photometa/internal/port"
)

// Tag represents a single IPTC metadata tag.
type Tag struct {
	record  uint8
	dataset uint8
	value   any
}

// Ensure Tag implements format.Tag interface.
var _ format.Tag = Tag{}

func (t Tag) Source() format.Source {
	return format.SourceIPTC
}

func (t Tag) ID() string {
	return fmt.Sprintf("%d:%d", t.record, t.dataset)
}

// Name returns the localized tag name using the current global locale.
func (t Tag) Name() string {
	return getLocalizedName(t.SpecName())
}

// SpecName returns the canonical English name from the IPTC-IIM specification.
func (t Tag) SpecName() string {
	return getSpecName(t.record, t.dataset)
}

func (t Tag) RawValue() any {
	return t.value
}

func (t Tag) String() string {
	return fmt.Sprintf("%v", t.value)
}

// Decode parses IPTC data from a byte slice.
// The data should be the raw IPTC IIM records (starting with 0x1C).
func Decode(data []byte, logger port.Logger) ([]Tag, error) {
	var tags []Tag
	offset := 0

	for offset < len(data) {
		// Valid IPTC record starts with 0x1C
		if data[offset] != 0x1C {
			offset++
			continue
		}

		if offset+5 > len(data) {
			if logger != nil {
				logger.Error("IPTC data too short for header", "offset", offset)
			}
			break
		}

		record := data[offset+1]
		dataset := data[offset+2]

		size := int(binary.BigEndian.Uint16(data[offset+3 : offset+5]))
		headerSize := 5

		// Extended Dataset (size > 32767) - MSB of first size byte is set
		if size&0x8000 != 0 {
			numLenBytes := size & 0x7FFF
			if offset+5+numLenBytes > len(data) {
				err := fmt.Errorf("truncated extended size")
				if logger != nil {
					logger.Error("IPTC parsing error", "error", err)
				}
				return tags, err
			}
			var extSize uint32
			for i := 0; i < numLenBytes && i < 4; i++ {
				extSize = (extSize << 8) | uint32(data[offset+5+i])
			}
			size = int(extSize)
			headerSize = 5 + numLenBytes
		}

		if offset+headerSize+size > len(data) {
			err := fmt.Errorf("truncated dataset value")
			if logger != nil {
				logger.Error("IPTC parsing error", "error", err)
			}
			return tags, err
		}

		valData := data[offset+headerSize : offset+headerSize+size]

		strVal := string(valData)
		strVal = strings.TrimRight(strVal, "\x00")

		tags = append(tags, Tag{
			record:  record,
			dataset: dataset,
			value:   strVal,
		})

		offset += headerSize + size
	}

	return tags, nil
}
