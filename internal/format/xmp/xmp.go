package xmp

import (
	"bytes"
	"encoding/xml"
	"fmt"

	"github.com/DementorAK/photometa/internal/format"
)

// Tag represents a single XMP metadata tag.
type Tag struct {
	namespace string
	name      string
	value     any
}

// Ensure Tag implements format.Tag interface.
var _ format.Tag = Tag{}

func (t Tag) Source() format.Source {
	return format.SourceXMP
}

func (t Tag) ID() string {
	return fmt.Sprintf("%s:%s", t.namespace, t.name)
}

// Name returns the localized tag name using the current global locale.
func (t Tag) Name() string {
	return getLocalizedName(t.SpecName())
}

// SpecName returns the canonical English name from the XMP specification.
func (t Tag) SpecName() string {
	return getSpecName(t.name)
}

func (t Tag) RawValue() any {
	return t.value
}

func (t Tag) String() string {
	return fmt.Sprintf("%v", t.value)
}

// xmpMeta represents the root of the XMP data structure we care about.
// We are looking for rdf:Description elements which contain the properties.
type rdfDescription struct {
	XMLName xml.Name
	// Due to varying structural forms of XMP, we parse attributes and inner XML separately
	Attrs []xml.Attr `xml:",any,attr"`
	Items []xmlItem  `xml:",any"`
}

type xmlItem struct {
	XMLName xml.Name
	Content []byte     `xml:",innerxml"`
	Attrs   []xml.Attr `xml:",any,attr"`
}

type rdfRDF struct {
	Descriptions []rdfDescription `xml:"Description"`
}

type xmpMetaRoot struct {
	RDF rdfRDF `xml:"RDF"`
}

// Decode parses XMP data from a byte slice containing the XML structure.
func Decode(data []byte) ([]Tag, error) {
	// Look for the <?xpacket begin=... ?> or <x:xmpmeta> prefix
	startIdx := bytes.Index(data, []byte("<x:xmpmeta"))
	if startIdx == -1 {
		// Sometimes it's directly <rdf:RDF>
		startIdx = bytes.Index(data, []byte("<rdf:RDF"))
		if startIdx == -1 {
			return nil, fmt.Errorf("no XMP metadata found")
		}
	}

	endIdx := bytes.LastIndex(data, []byte("xmpmeta>"))
	if endIdx != -1 {
		endIdx += len("xmpmeta>")
	} else {
		endIdx = bytes.LastIndex(data, []byte("RDF>"))
		if endIdx != -1 {
			endIdx += len("RDF>")
		}
	}

	if endIdx == -1 || endIdx <= startIdx {
		startIdx = 0
		endIdx = len(data)
	}

	xmlData := data[startIdx:endIdx]

	decoder := xml.NewDecoder(bytes.NewReader(xmlData))

	var tags []Tag

	for {
		t, err := decoder.Token()
		if err != nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "Description" {
				// Parse attributes on rdf:Description which are effectively properties
				for _, attr := range se.Attr {
					// skip xmlns definitions and rdf:about
					if attr.Name.Space == "xmlns" || attr.Name.Space == "http://www.w3.org/1999/02/22-rdf-syntax-ns#" && attr.Name.Local == "about" {
						continue
					}
					if attr.Name.Space == "xml" {
						continue
					}

					tags = append(tags, Tag{
						namespace: "attr",
						name:      attr.Name.Local,
						value:     attr.Value,
					})
				}

				// Parse inner elements of Description
				var description rdfDescription
				decoder.DecodeElement(&description, &se)

				for _, item := range description.Items {
					val := string(item.Content)
					if bytes.Contains(item.Content, []byte("<rdf:Alt>")) || bytes.Contains(item.Content, []byte("<rdf:Seq>")) || bytes.Contains(item.Content, []byte("<rdf:Bag>")) {
						val = extractBagElements(item.Content)
					} else {
						val = extractTextOnly(item.Content)
					}

					tags = append(tags, Tag{
						namespace: "xml",
						name:      item.XMLName.Local,
						value:     val,
					})
				}
			}
		}
	}

	return tags, nil
}

func extractBagElements(data []byte) string {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var items []string
	for {
		t, err := decoder.Token()
		if err != nil {
			break
		}
		if se, ok := t.(xml.StartElement); ok && se.Name.Local == "li" {
			var val string
			decoder.DecodeElement(&val, &se)
			if val != "" {
				items = append(items, val)
			}
		}
	}
	result := ""
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		result += item
	}
	return result
}

func extractTextOnly(data []byte) string {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var textData string
	for {
		t, err := decoder.Token()
		if err != nil {
			break
		}
		if cd, ok := t.(xml.CharData); ok {
			textData += string(cd)
		}
	}
	return textData
}
