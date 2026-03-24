package filler

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/DementorAK/photometa/internal/format/imageformat"
	"github.com/DementorAK/photometa/internal/port"
)

type segments struct {
	exif          []byte
	iptc          []byte
	xmp           []byte
	width, height int
}

// extractSegments parses the image structure to find metadata blobs and dimensions.
func extractSegments(r io.ReadSeeker, format imageformat.Format, logger port.Logger) (segments, error) {
	var segs segments

	// Reset position to start
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return segs, err
	}

	switch format {
	case imageformat.JPEG:
		return extractSegmentsJPEG(r)
	case imageformat.PNG:
		return extractSegmentsPNG(r, logger)
	case imageformat.WebP:
		return extractSegmentsWebP(r, logger)
	case imageformat.TIFF:
		return extractSegmentsTIFF(r)
	default:
		logger.Warn("metadata extraction not implemented for format", "format", format)
		return segs, fmt.Errorf("metadata extraction not implemented for format %s", format)
	}
}

func extractSegmentsJPEG(r io.ReadSeeker) (segments, error) {
	var segs segments

	// Read SOI
	magic := make([]byte, 2)
	if _, err := io.ReadFull(r, magic); err != nil {
		return segs, err
	}
	if magic[0] != 0xFF || magic[1] != 0xD8 {
		return segs, fmt.Errorf("not a JPEG")
	}

	for {
		// Read marker
		buf := make([]byte, 1)

		// 1. Wait for at least one 0xFF
		for {
			if _, err := io.ReadFull(r, buf); err != nil {
				return segs, nil
			}
			if buf[0] == 0xFF {
				break
			}
		}

		// 2. Skip all subsequent 0xFFs
		var m byte
		for {
			if _, err := io.ReadFull(r, buf); err != nil {
				return segs, nil
			}
			m = buf[0]
			if m != 0xFF {
				break
			}
		}

		if m == 0x00 {
			break // loss of sync or byte stuffing
		}

		if m == 0xD9 || m == 0xDA { // EOI or SOS, stop parsing segments
			break
		}

		// Read length
		lenBytes := make([]byte, 2)
		if _, err := io.ReadFull(r, lenBytes); err != nil {
			break
		}
		length := int(binary.BigEndian.Uint16(lenBytes)) - 2
		if length < 0 {
			break
		}

		// SOF0 - SOF15 (except DHT, JPG, DAC) for dimensions
		if (m >= 0xC0 && m <= 0xCF) && m != 0xC4 && m != 0xC8 && m != 0xCC {
			if length >= 5 {
				data := make([]byte, 5)
				if _, err := io.ReadFull(r, data); err != nil {
					break
				}
				segs.height = int(binary.BigEndian.Uint16(data[1:3]))
				segs.width = int(binary.BigEndian.Uint16(data[3:5]))
				if _, err := r.Seek(int64(length-5), io.SeekCurrent); err != nil {
					break
				}
			} else {
				if _, err := r.Seek(int64(length), io.SeekCurrent); err != nil {
					break
				}
			}
			continue
		}

		// APP1 (Exif / XMP)
		if m == 0xE1 {
			data := make([]byte, length)
			if _, err := io.ReadFull(r, data); err != nil {
				break
			}

			if len(data) >= 6 && bytes.HasPrefix(data, []byte("Exif\x00\x00")) {
				if len(segs.exif) == 0 {
					segs.exif = data
				} else {
					segs.exif = append(segs.exif, data...)
				}
			} else if len(data) >= 29 && bytes.HasPrefix(data, []byte("http://ns.adobe.com/xap/1.0/\x00")) {
				if len(segs.xmp) == 0 {
					segs.xmp = data
				} else {
					segs.xmp = append(segs.xmp, data...)
				}
			} else if len(data) >= 35 && bytes.HasPrefix(data, []byte("http://ns.adobe.com/xmp/extension/\x00")) {
				segs.xmp = append(segs.xmp, data...)
			}
			continue
		}

		// APP13 (IPTC)
		if m == 0xED {
			data := make([]byte, length)
			if _, err := io.ReadFull(r, data); err != nil {
				break
			}

			if len(data) >= 1 && data[0] == 0x1C {
				segs.iptc = data
			} else if len(data) >= 14 && bytes.HasPrefix(data, []byte("Photoshop 3.0\x00")) {
				if parsed := extractIPTCFromIRB(data[14:]); len(parsed) > 0 {
					segs.iptc = parsed
				}
			} else if len(data) >= 20 && bytes.HasPrefix(data, []byte("Adobe_Photoshop2.0\x00")) {
				if parsed := extractIPTCFromIRB(data[20:]); len(parsed) > 0 {
					segs.iptc = parsed
				}
			}
			continue
		}

		// Skip other segments
		if _, err := r.Seek(int64(length), io.SeekCurrent); err != nil {
			break
		}
	}

	return segs, nil
}

func extractIPTCFromIRB(data []byte) []byte {
	offset := 0
	for offset+10 < len(data) {
		if string(data[offset:offset+4]) != "8BIM" {
			offset++
			continue
		}
		id := binary.BigEndian.Uint16(data[offset+4 : offset+6])
		nameLen := int(data[offset+6])
		nameLenTotal := (nameLen + 2) &^ 1 // total length including the nameLen byte and padding

		if offset+6+nameLenTotal+4 > len(data) {
			break
		}

		sizeOffset := offset + 6 + nameLenTotal
		size := int(binary.BigEndian.Uint32(data[sizeOffset : sizeOffset+4]))
		sizePadded := (size + 1) &^ 1

		if id == 0x0404 { // IPTC NAAC
			dataOffset := sizeOffset + 4
			if dataOffset+size <= len(data) {
				return data[dataOffset : dataOffset+size]
			}
		}

		offset = sizeOffset + 4 + sizePadded
	}
	return nil
}

func extractSegmentsPNG(r io.ReadSeeker, logger port.Logger) (segments, error) {
	var segs segments

	// PNG signature
	sig := make([]byte, 8)
	if _, err := io.ReadFull(r, sig); err != nil {
		return segs, err
	}
	if !bytes.Equal(sig, []byte("\x89PNG\r\n\x1a\n")) {
		return segs, fmt.Errorf("not a PNG")
	}

	for {
		// Read chunk length
		lenBytes := make([]byte, 4)
		if _, err := io.ReadFull(r, lenBytes); err != nil {
			if err == io.EOF {
				break
			}
			return segs, err
		}
		length := binary.BigEndian.Uint32(lenBytes)

		// Read chunk type
		typeBytes := make([]byte, 4)
		if _, err := io.ReadFull(r, typeBytes); err != nil {
			return segs, err
		}
		chunkType := string(typeBytes)

		// Read chunk data
		data := make([]byte, length)
		if _, err := io.ReadFull(r, data); err != nil {
			return segs, err
		}

		// Read CRC
		crc := make([]byte, 4)
		if _, err := io.ReadFull(r, crc); err != nil {
			return segs, err
		}

		switch chunkType {
		case "IHDR":
			if length >= 8 {
				segs.width = int(binary.BigEndian.Uint32(data[0:4]))
				segs.height = int(binary.BigEndian.Uint32(data[4:8]))
			}
		case "eXIf":
			// Raw EXIF data
			segs.exif = data
		case "iTXt", "zTXt", "tEXt":
			// Sometimes XMP is in generic text chunks
			if bytes.HasPrefix(data, []byte("XML:com.adobe.xmp\x00")) {
				parts := bytes.SplitN(data, []byte("\x00"), 3) // Need to handle compression byte if zTXt
				if len(parts) > 1 {
					// Extremely simplified, proper extraction requires zlib decompression for zTXt/iTXt
					segs.xmp = data
				}
			}
			logger.Debug("extracted text chunk from PNG", "type", chunkType, "len", length)
		case "IEND":
			return segs, nil
		}
	}

	return segs, nil
}

func extractSegmentsWebP(r io.ReadSeeker, logger port.Logger) (segments, error) {
	var segs segments

	// RIFF signature
	sig := make([]byte, 12)
	if _, err := io.ReadFull(r, sig); err != nil {
		return segs, err
	}
	if !bytes.Equal(sig[0:4], []byte("RIFF")) || !bytes.Equal(sig[8:12], []byte("WEBP")) {
		return segs, fmt.Errorf("not a WebP")
	}

	for {
		header := make([]byte, 8)
		if _, err := io.ReadFull(r, header); err != nil {
			if err == io.EOF {
				break
			}
			return segs, err
		}

		chunkType := string(header[0:4])
		// WebP chunk sizes are little-endian
		length := binary.LittleEndian.Uint32(header[4:8])
		// Chunk data must be even length, padded with null byte if odd
		paddedLen := length
		if length%2 != 0 {
			paddedLen++
		}

		switch chunkType {
		case "VP8X":
			if length >= 10 {
				data := make([]byte, 10)
				if _, err := io.ReadFull(r, data); err != nil {
					return segs, err
				}
				// 24-bit little endian width and height
				segs.width = 1 + int(data[4]) | int(data[5])<<8 | int(data[6])<<16
				segs.height = 1 + int(data[7]) | int(data[8])<<8 | int(data[9])<<16
				r.Seek(int64(paddedLen-10), io.SeekCurrent)
			} else {
				r.Seek(int64(paddedLen), io.SeekCurrent)
			}
		case "EXIF":
			segs.exif = make([]byte, length)
			if _, err := io.ReadFull(r, segs.exif); err != nil {
				return segs, err
			}
			if paddedLen > length {
				r.Seek(1, io.SeekCurrent)
			}
			logger.Debug("extracted EXIF from WebP", "len", length)
		case "XMP ":
			segs.xmp = make([]byte, length)
			if _, err := io.ReadFull(r, segs.xmp); err != nil {
				return segs, err
			}
			if paddedLen > length {
				r.Seek(1, io.SeekCurrent)
			}
			logger.Debug("extracted XMP from WebP", "len", length)
		default:
			r.Seek(int64(paddedLen), io.SeekCurrent)
		}
	}

	return segs, nil
}

func extractSegmentsTIFF(r io.ReadSeeker) (segments, error) {
	var segs segments
	// For TIFF, the whole file is the EXIF container essentially.
	data, _ := io.ReadAll(r)
	segs.exif = data
	return segs, nil
}
