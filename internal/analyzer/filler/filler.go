package filler

import (
	"io"
	"math"

	"github.com/DementorAK/photometa/internal/domain"
	"github.com/DementorAK/photometa/internal/format/exif"
	"github.com/DementorAK/photometa/internal/format/imageformat"
	"github.com/DementorAK/photometa/internal/format/iptc"
	"github.com/DementorAK/photometa/internal/format/xmp"
	"github.com/DementorAK/photometa/internal/port"
)

// Fill extracts EXIF, IPTC, and XMP metadata using custom parsers and fills domain.Metadata.
func Fill(r io.ReadSeeker, meta *domain.Metadata, logger port.Logger) {
	format := imageformat.Format(meta.Format)
	segs, err := extractSegments(r, format, logger)
	if err != nil {
		logger.Warn("metadata extraction error", "format", format, "error", err)
		return
	}

	// Fill image dimensions from decode result
	if segs.width > 0 {
		addTag(meta, "Format", "Photo", "Width", segs.width)
	}
	if segs.height > 0 {
		addTag(meta, "Format", "Photo", "Height", segs.height)
	}

	// Process EXIF tags
	if len(segs.exif) > 0 {
		exifTags, err := exif.Decode(segs.exif)
		if err == nil {
			fillExifTags(exifTags, meta, logger)
		}
	}

	// Process IPTC tags
	if len(segs.iptc) > 0 {
		iptcTags, err := iptc.Decode(segs.iptc)
		if err == nil {
			fillIPTCTags(iptcTags, meta, logger)
		}
	}

	// Process XMP tags
	if len(segs.xmp) > 0 {
		xmpTags, err := xmp.Decode(segs.xmp)
		if err == nil {
			fillXMPTags(xmpTags, meta, logger)
		}
	}
}

func addTag(meta *domain.Metadata, tType, group, name string, value any) {
	if value == nil {
		return
	}
	if s, ok := value.(string); ok && s == "" {
		return
	}
	meta.Tags = append(meta.Tags, domain.TagInfo{
		Type:  tType,
		Group: group,
		Name:  name,
		Value: value,
	})
}

func getTagValue(meta *domain.Metadata, name string) any {
	for _, t := range meta.Tags {
		if t.Name == name {
			return t.Value
		}
	}
	return nil
}

func updateGPSRef(meta *domain.Metadata, tagName string, multiplier float64) {
	for i, t := range meta.Tags {
		if t.Group == "Location" && t.Name == tagName {
			if val, ok := t.Value.(float64); ok && val > 0 {
				meta.Tags[i].Value = val * multiplier
			}
			break
		}
	}
}

func fillExifTags(tags []exif.Tag, meta *domain.Metadata, logger port.Logger) {
	for _, tag := range tags {
		name := tag.Name()
		val := tag.RawValue()
		strVal := tag.String()

		switch name {
		// Equipment
		case "Make", "Model", "Software":
			addTag(meta, "EXIF", "Equipment", name, strVal)
		case "LensModel":
			addTag(meta, "EXIF", "Equipment", "LensInfo", strVal)

		// Shooting params
		case "ExposureTime", "FocalLength", "Flash", "WhiteBalance":
			addTag(meta, "EXIF", "Shooting", name, strVal)
		case "FNumber":
			addTag(meta, "EXIF", "Shooting", name, TagToFloat(val))
		case "ISOSpeedRatings":
			addTag(meta, "EXIF", "Shooting", "ISO", TagToInt(val))

		// Photo properties
		case "Orientation":
			addTag(meta, "EXIF", "Photo", name, TagToInt(val))
		case "ColorSpace":
			addTag(meta, "EXIF", "Photo", name, strVal)

		// Date/Location
		case "DateTimeOriginal", "Create Date":
			if t := TagToTime(val, strVal); t != nil {
				addTag(meta, "EXIF", "Location", "DateTaken", t)
			}
		case "GPSLatitude", "GPSLongitude":
			addTag(meta, "EXIF", "Location", name, TagToFloat(val))
		case "GPSLatitudeRef":
			if strVal == "S" || strVal == "S\x00" {
				updateGPSRef(meta, "GPSLatitude", -1.0)
			} else {
				addTag(meta, "EXIF", "Other", name, strVal)
			}
		case "GPSLongitudeRef":
			if strVal == "W" || strVal == "W\x00" {
				updateGPSRef(meta, "GPSLongitude", -1.0)
			} else {
				addTag(meta, "EXIF", "Other", name, strVal)
			}

		// Author
		case "Artist", "Copyright":
			addTag(meta, "EXIF", "Author", name, strVal)

		default:
			// Resolve deferred GPS refs if late
			if name == "GPSLatitude" {
				if ref, ok := getTagValue(meta, "GPSLatitudeRef").(string); ok && (ref == "S" || ref == "S\x00") {
					valF := TagToFloat(val)
					addTag(meta, "EXIF", "Location", name, -math.Abs(valF))
					continue
				}
			}
			if name == "GPSLongitude" {
				if ref, ok := getTagValue(meta, "GPSLongitudeRef").(string); ok && (ref == "W" || ref == "W\x00") {
					valF := TagToFloat(val)
					addTag(meta, "EXIF", "Location", name, -math.Abs(valF))
					continue
				}
			}
			logger.Debug("unhandled EXIF tag", "tag", name, "value", strVal)
			addTag(meta, "EXIF", "Other", name, strVal)
		}
	}
}

func fillIPTCTags(tags []iptc.Tag, meta *domain.Metadata, logger port.Logger) {
	for _, tag := range tags {
		name := tag.Name()
		strVal := tag.String()

		switch name {
		case "By-line":
			addTag(meta, "IPTC", "Author", "Creator", strVal)
		case "Credit":
			addTag(meta, "IPTC", "Author", "Credit", strVal)
		case "CopyrightNotice":
			addTag(meta, "IPTC", "Author", "Rights", strVal)
		default:
			logger.Debug("unhandled IPTC tag", "tag", name, "value", strVal)
			addTag(meta, "IPTC", "Other", name, strVal)
		}
	}
}

func fillXMPTags(tags []xmp.Tag, meta *domain.Metadata, logger port.Logger) {
	for _, tag := range tags {
		name := tag.Name()
		strVal := tag.String()

		switch name {
		case "Rights":
			addTag(meta, "XMP", "Author", "Rights", strVal)
		case "Creator":
			addTag(meta, "XMP", "Author", "Creator", strVal)
		default:
			logger.Debug("unhandled XMP tag", "tag", name, "value", strVal)
			addTag(meta, "XMP", "Other", name, strVal)
		}
	}
}
