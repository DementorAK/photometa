package locale

import (
	"testing"
)

func TestDefaultLocale(t *testing.T) {
	// Reset to default for this test
	mu.Lock()
	current = "en"
	mu.Unlock()

	if got := Locale(); got != "en" {
		t.Errorf("expected default locale 'en', got %q", got)
	}
}

func TestSetLocale_ValidCode(t *testing.T) {
	defer func() {
		mu.Lock()
		current = "en"
		mu.Unlock()
	}()

	codes := []string{"ua", "ru", "de", "fr", "es", "en"}
	for _, code := range codes {
		SetLocale(code)
		if got := Locale(); got != code {
			t.Errorf("SetLocale(%q): expected %q, got %q", code, code, got)
		}
	}
}

func TestSetLocale_InvalidCode(t *testing.T) {
	defer func() {
		mu.Lock()
		current = "en"
		mu.Unlock()
	}()

	SetLocale("en")
	SetLocale("xx")
	if got := Locale(); got != "en" {
		t.Errorf("expected locale to remain 'en' after invalid code, got %q", got)
	}
}

func TestSetLocale_EmptyCode(t *testing.T) {
	defer func() {
		mu.Lock()
		current = "en"
		mu.Unlock()
	}()

	SetLocale("ua")
	SetLocale("")
	if got := Locale(); got != "ua" {
		t.Errorf("expected locale to remain 'ua' after empty code, got %q", got)
	}
}

func TestGetLocales_ReturnsAllLocales(t *testing.T) {
	locs := GetLocales()

	if len(locs) != 6 {
		t.Fatalf("expected 6 locales, got %d", len(locs))
	}

	expectedCodes := map[string]bool{
		"en": false, "ua": false, "ru": false,
		"de": false, "fr": false, "es": false,
	}

	for _, loc := range locs {
		if _, ok := expectedCodes[loc.Code]; !ok {
			t.Errorf("unexpected locale code %q", loc.Code)
		}
		expectedCodes[loc.Code] = true

		if loc.Description == "" {
			t.Errorf("locale %q has empty description", loc.Code)
		}
	}

	for code, found := range expectedCodes {
		if !found {
			t.Errorf("missing expected locale %q", code)
		}
	}
}

func TestGetLocales_ReturnsCopy(t *testing.T) {
	locs := GetLocales()
	locs[0].Code = "xx"

	original := GetLocales()
	if original[0].Code == "xx" {
		t.Error("GetLocales should return a copy, but original was modified")
	}
}

func TestT_EnglishFallback(t *testing.T) {
	defer func() {
		mu.Lock()
		current = "en"
		mu.Unlock()
	}()

	SetLocale("en")
	if got := T("Size"); got != "Size" {
		t.Errorf("expected 'Size', got %q", got)
	}
}

func TestT_RussianTranslation(t *testing.T) {
	defer func() {
		mu.Lock()
		current = "en"
		mu.Unlock()
	}()

	SetLocale("ua")
	if got := T("Size"); got != "Розмір" {
		t.Errorf("expected 'Розмір', got %q", got)
	}
}

func TestT_UnknownKey(t *testing.T) {
	defer func() {
		mu.Lock()
		current = "en"
		mu.Unlock()
	}()

	SetLocale("ua")
	if got := T("unknown_key"); got != "unknown_key" {
		t.Errorf("expected 'unknown_key', got %q", got)
	}
}
