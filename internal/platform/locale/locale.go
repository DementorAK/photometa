// Package locale provides global locale management for the entire application.
//
// It defines the set of available locales, allows switching the active locale
// via [SetLocale], and provides [GetLocales] for UI elements (dropdowns, CLI flags, HTTP endpoints).
//
// By default the locale is "en" (English).
package locale

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed locales/*.json
var localesFS embed.FS

var (
	translations     map[string]map[string]string
	translationsOnce sync.Once
	translationsMu   sync.RWMutex
)

// Info describes a single available locale.
type Info struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// locales is the registry of all supported locales.
// Descriptions are given in the respective language (standard UX for language selectors).
var locales = []Info{
	{Code: "en", Description: "English"},
	{Code: "ua", Description: "Українська"},
	{Code: "ru", Description: "Русский"},
	{Code: "es", Description: "Español"},
	{Code: "de", Description: "Deutsch"},
	{Code: "fr", Description: "Français"},
}

var (
	current = "en"
	mu      sync.RWMutex
)

// SetLocale sets the global locale for all display strings.
// The code must match one of the codes in [GetLocales].
// If code is empty or unknown, the locale is not changed.
func SetLocale(code string) {
	if code == "" {
		return
	}
	for _, loc := range locales {
		if loc.Code == code {
			mu.Lock()
			current = code
			mu.Unlock()
			return
		}
	}
}

// Locale returns the current global locale code (e.g. "en", "ua", "ru", "es", "de", "fr").
func Locale() string {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// GetLocales returns a copy of all available locales.
func GetLocales() []Info {
	result := make([]Info, len(locales))
	copy(result, locales)
	return result
}

// T returns the translation of key for the current locale.
// If no translation exists, returns the key itself (English fallback).
func T(key string) string {
	mu.RLock()
	lang := current
	mu.RUnlock()

	return TForLocale(lang, key)
}

// TForLocale returns the translation of key for a specific locale code.
// If no translation exists or code is "en", returns the key itself.
func TForLocale(code, key string) string {
	if code == "en" {
		return key
	}

	translationsOnce.Do(loadTranslations)

	translationsMu.RLock()
	defer translationsMu.RUnlock()

	if dict, ok := translations[code]; ok {
		if val, ok := dict[key]; ok {
			return val
		}
	}
	return key
}

func loadTranslations() {
	translationsMu.Lock()
	defer translationsMu.Unlock()

	translations = make(map[string]map[string]string)
	for _, loc := range locales {
		if loc.Code == "en" {
			continue
		}
		data, err := localesFS.ReadFile(fmt.Sprintf("locales/%s.json", loc.Code))
		if err == nil {
			var m map[string]string
			if json.Unmarshal(data, &m) == nil {
				translations[loc.Code] = m
			}
		}
	}
}
