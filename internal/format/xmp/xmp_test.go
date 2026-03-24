package xmp

import (
	"testing"

	"github.com/DementorAK/photometa/internal/platform/locale"
)

func TestDecode(t *testing.T) {
	xmlData := []byte(`
<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description rdf:about=""
      xmlns:xmp="http://ns.adobe.com/xap/1.0/"
      xmlns:dc="http://purl.org/dc/elements/1.1/"
      xmlns:exif="http://ns.adobe.com/exif/1.0/"
      xmp:CreatorTool="Adobe Photoshop"
      xmp:CreateDate="2023-10-01T12:00:00">
      <dc:creator>
        <rdf:Seq>
          <rdf:li>John Doe</rdf:li>
          <rdf:li>Jane Doe</rdf:li>
        </rdf:Seq>
      </dc:creator>
      <dc:title>
        <rdf:Alt>
          <rdf:li xml:lang="x-default">Sample Photo</rdf:li>
        </rdf:Alt>
      </dc:title>
      <exif:PixelXDimension>1920</exif:PixelXDimension>
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>
	`)

	tags, err := Decode(xmlData)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(tags) != 5 {
		t.Fatalf("expected 5 tags, got %d", len(tags))
	}

	foundCreatorTool := false
	foundCreator := false
	for _, tag := range tags {
		if tag.name == "CreatorTool" {
			foundCreatorTool = true
			if tag.RawValue() != "Adobe Photoshop" {
				t.Errorf("CreatorTool value = %v, want Adobe Photoshop", tag.RawValue())
			}
			if tag.SpecName() != "CreatorTool" {
				t.Errorf("CreatorTool SpecName = %s, want CreatorTool", tag.SpecName())
			}
		}
		if tag.name == "creator" {
			foundCreator = true
			if tag.RawValue() != "John Doe, Jane Doe" {
				t.Errorf("creator value = %v, want 'John Doe, Jane Doe'", tag.RawValue())
			}
		}
	}

	if !foundCreatorTool || !foundCreator {
		t.Error("Missing expected properties")
	}
}

func TestXMP_RussianLocale(t *testing.T) {
	locale.SetLocale("ru")
	defer locale.SetLocale("en")

	tag := Tag{namespace: "dc", name: "creator"}
	if name := tag.Name(); name != "Автор" {
		t.Errorf("Name(ru) = %s, want Автор", name)
	}
	if name := tag.SpecName(); name != "Creator" {
		t.Errorf("SpecName = %s, want Creator", name)
	}
}

func TestXMP_DefaultLocale(t *testing.T) {
	locale.SetLocale("en")
	tag := Tag{namespace: "dc", name: "creator"}
	if name := tag.Name(); name != "Creator" {
		t.Errorf("Name(en) = %s, want Creator", name)
	}
}

func TestDecode_NoXMP(t *testing.T) {
	_, err := Decode([]byte("Some random binary data that does not have xmpmeta or rdf:RDF"))
	if err == nil {
		t.Error("expected error for no XMP present")
	}
}
