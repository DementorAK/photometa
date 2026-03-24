package exif

import (
	"embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/DementorAK/photometa/internal/platform/locale"
)

//go:embed fields.json
var fieldsData []byte

//go:embed gps_fields.json
var gpsFieldsData []byte

//go:embed locales
var localesFS embed.FS

// fieldDef represents a tag definition from the JSON registry.
type fieldDef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var (
	// fieldsMap maps tag ID → spec name for IFD0/ExifIFD tags
	fieldsMap map[uint16]string
	// gpsFieldsMap maps tag ID → spec name for GPS IFD tags
	gpsFieldsMap map[uint16]string
	fieldsOnce   sync.Once

	// localeCache caches loaded locale translations: lang → (specName → translation)
	localeCache map[string]map[string]string
	localeMu    sync.Mutex
)

// ensureFieldsLoaded loads the tag registries from embedded JSON files once.
func ensureFieldsLoaded() {
	fieldsOnce.Do(func() {
		fieldsMap = loadFieldMap(fieldsData)
		gpsFieldsMap = loadFieldMap(gpsFieldsData)
		localeCache = make(map[string]map[string]string)
	})
}

// loadFieldMap parses a JSON array of fieldDef into a map[uint16]string.
func loadFieldMap(data []byte) map[uint16]string {
	var fields []fieldDef
	if err := json.Unmarshal(data, &fields); err != nil {
		panic(fmt.Errorf("failed to parse embedded fields JSON: %w", err))
	}

	m := make(map[uint16]string, len(fields))
	for _, f := range fields {
		hexStr := strings.TrimPrefix(f.ID, "0x")
		id64, err := strconv.ParseUint(hexStr, 16, 16)
		if err != nil {
			continue
		}
		m[uint16(id64)] = f.Name
	}
	return m
}

// getLocaleMap returns the translation map for the given language,
// loading it from the embedded locales directory if needed.
func getLocaleMap(lang string) map[string]string {
	if lang == "" || lang == "en" {
		return nil // English names come from the spec, no translation needed
	}

	localeMu.Lock()
	defer localeMu.Unlock()

	if m, ok := localeCache[lang]; ok {
		return m
	}

	// Try to load locales/{lang}.json
	filename := fmt.Sprintf("locales/%s.json", lang)
	data, err := localesFS.ReadFile(filename)
	if err != nil {
		// No translation file for this language — cache nil
		localeCache[lang] = nil
		return nil
	}

	m := make(map[string]string)
	if err := json.Unmarshal(data, &m); err != nil {
		localeCache[lang] = nil
		return nil
	}

	localeCache[lang] = m
	return m
}

// getSpecName returns the specification name for a tag ID.
// It looks up in the main fields map first, then GPS fields.
// If unknown, returns "Unknown_0xNNNN".
func getSpecName(id uint16, isGPS bool) string {
	ensureFieldsLoaded()

	if isGPS {
		if name, ok := gpsFieldsMap[id]; ok {
			return name
		}
	} else {
		if name, ok := fieldsMap[id]; ok {
			return name
		}
	}

	return fmt.Sprintf("Unknown_0x%04X", id)
}

// getLocalizedName returns the translated name for a tag, using the
// current global locale. Falls back to the spec name.
func getLocalizedName(specName string) string {
	lang := locale.Locale()
	if lang == "" || lang == "en" {
		return specName
	}

	m := getLocaleMap(lang)
	if m == nil {
		return specName
	}

	if translated, ok := m[specName]; ok {
		return translated
	}

	return specName
}
