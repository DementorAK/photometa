package xmp

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/DementorAK/photometa/internal/platform/locale"
)

//go:embed fields.json
var fieldsData []byte

//go:embed locales
var localesFS embed.FS

type fieldDef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var (
	fieldsMap  map[string]string // id → spec name
	fieldsOnce sync.Once

	localeCache map[string]map[string]string
	localeMu    sync.Mutex
)

func ensureFieldsLoaded() {
	fieldsOnce.Do(func() {
		var fields []fieldDef
		if err := json.Unmarshal(fieldsData, &fields); err != nil {
			panic(fmt.Errorf("failed to parse embedded XMP fields.json: %w", err))
		}

		fieldsMap = make(map[string]string, len(fields))
		for _, f := range fields {
			fieldsMap[f.ID] = f.Name
		}

		localeCache = make(map[string]map[string]string)
	})
}

func getLocaleMap(lang string) map[string]string {
	if lang == "" || lang == "en" {
		return nil
	}

	localeMu.Lock()
	defer localeMu.Unlock()

	if m, ok := localeCache[lang]; ok {
		return m
	}

	filename := fmt.Sprintf("locales/%s.json", lang)
	data, err := localesFS.ReadFile(filename)
	if err != nil {
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

// getSpecName returns the canonical name for an XMP property.
// For XMP, the local name from XML is used as lookup key.
// If not found in registry, the original name is returned as-is.
func getSpecName(name string) string {
	ensureFieldsLoaded()
	if specName, ok := fieldsMap[name]; ok {
		return specName
	}
	return name
}

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
