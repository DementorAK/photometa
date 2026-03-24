package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DementorAK/photometa/internal/domain"
	"github.com/DementorAK/photometa/internal/fake"
)

// ============================================================================
// HTTP SERVER TESTS
// ============================================================================

func TestHandleAnalyze_Success(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	mockAnalyzer.AnalyzeStreamFn = func(ctx context.Context, r io.Reader, name string, size int64) (*domain.ImageFile, error) {
		return &domain.ImageFile{
			Name: name,
			Metadata: domain.Metadata{
				FileSize: size,
				Format:   "jpeg",
				Tags: []domain.TagInfo{
					{Type: "EXIF", Group: "Equipment", Name: "Make", Value: "Canon"},
					{Type: "EXIF", Group: "Equipment", Name: "Model", Value: "EOS 5D"},
				},
			},
		}, nil
	}

	server := NewServer(mockAnalyzer)

	// Create multipart form request
	body, contentType := createMultipartRequest(t, "file", "test.jpg", []byte("fake jpeg data"))
	req := httptest.NewRequest(http.MethodPost, "/analyze", body)
	req.Header.Set("Content-Type", contentType)

	rec := httptest.NewRecorder()

	// Act
	server.handleAnalyze(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", rec.Header().Get("Content-Type"))
	}

	// Verify JSON response
	var result AnalyzeResponse
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify that the metadata contains the expected property
	found := false
	for _, p := range result.Metadata {
		if p.Synonym == "Make" && p.Value == "Canon" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected Equipment Make 'Canon' in metadata, got %v", result.Metadata)
	}

	// Verify the analyzer was called
	if len(mockAnalyzer.AnalyzeStreamCalls) != 1 {
		t.Errorf("expected 1 AnalyzeStream call, got %d", len(mockAnalyzer.AnalyzeStreamCalls))
	}

	if mockAnalyzer.AnalyzeStreamCalls[0].Name != "test.jpg" {
		t.Errorf("expected filename 'test.jpg', got %q", mockAnalyzer.AnalyzeStreamCalls[0].Name)
	}
}

func TestHandleAnalyze_WrongMethod(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	server := NewServer(mockAnalyzer)

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/analyze", nil)
			rec := httptest.NewRecorder()

			// Act
			server.handleAnalyze(rec, req)

			// Assert
			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status %d for %s, got %d", http.StatusMethodNotAllowed, method, rec.Code)
			}

			// Analyzer should NOT be called
			if len(mockAnalyzer.AnalyzeStreamCalls) != 0 {
				t.Errorf("analyzer should not be called for %s", method)
			}
		})
	}
}

func TestHandleAnalyze_MissingFile(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	server := NewServer(mockAnalyzer)

	req := httptest.NewRequest(http.MethodPost, "/analyze", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	rec := httptest.NewRecorder()

	// Act
	server.handleAnalyze(rec, req)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandleAnalyze_AnalyzerError(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	mockAnalyzer.AnalyzeStreamFn = func(ctx context.Context, r io.Reader, name string, size int64) (*domain.ImageFile, error) {
		return nil, errors.New("corrupted image file")
	}

	server := NewServer(mockAnalyzer)

	body, contentType := createMultipartRequest(t, "file", "corrupted.jpg", []byte("bad data"))
	req := httptest.NewRequest(http.MethodPost, "/analyze", body)
	req.Header.Set("Content-Type", contentType)

	rec := httptest.NewRecorder()

	// Act
	server.handleAnalyze(rec, req)

	// Assert
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	// Error message should be in response body
	if !bytes.Contains(rec.Body.Bytes(), []byte("corrupted image file")) {
		t.Errorf("expected error message in response, got: %s", rec.Body.String())
	}
}

// ============================================================================
// LOCALES ENDPOINT TESTS
// ============================================================================

func TestHandleLocales_Success(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	server := NewServer(mockAnalyzer)

	req := httptest.NewRequest(http.MethodGet, "/locales", nil)
	rec := httptest.NewRecorder()

	// Act
	server.handleLocales(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", rec.Header().Get("Content-Type"))
	}

	// Verify JSON response is an array of locale objects
	var result []struct {
		Code        string `json:"code"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("expected at least one locale, got empty list")
	}

	// Verify "en" locale is present
	found := false
	for _, loc := range result {
		if loc.Code == "en" {
			found = true
			if loc.Description != "English" {
				t.Errorf("expected description 'English' for 'en', got %q", loc.Description)
			}
			break
		}
	}
	if !found {
		t.Error("expected locale 'en' in response")
	}
}

func TestHandleLocales_WrongMethod(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	server := NewServer(mockAnalyzer)

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/locales", nil)
			rec := httptest.NewRecorder()

			// Act
			server.handleLocales(rec, req)

			// Assert
			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status %d for %s, got %d", http.StatusMethodNotAllowed, method, rec.Code)
			}
		})
	}
}

func TestHandleDemo_Success(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	server := NewServer(mockAnalyzer)

	req := httptest.NewRequest(http.MethodGet, "/demo", nil)
	rec := httptest.NewRecorder()

	// Act
	server.handleDemo(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Content-Type") != "text/html" {
		t.Errorf("expected Content-Type 'text/html', got %q", rec.Header().Get("Content-Type"))
	}

	if len(rec.Body.Bytes()) == 0 {
		t.Error("expected non-empty body for demo page")
	}
}

// ============================================================================
// TEST HELPERS
// ============================================================================

// createMultipartRequest creates a multipart form request with a file.
func createMultipartRequest(t *testing.T, fieldName, fileName string, content []byte) (*bytes.Buffer, string) {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	if _, err := part.Write(content); err != nil {
		t.Fatalf("failed to write file content: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	return &buf, writer.FormDataContentType()
}

func TestLocalizeMetadata(t *testing.T) {
	srv := NewServer(fake.NewMockImageAnalyzer())

	tests := []struct {
		name string
		meta domain.Metadata
		lang string
	}{
		{
			name: "empty metadata",
			meta: domain.Metadata{},
			lang: "en",
		},
		{
			name: "with equipment",
			meta: domain.Metadata{
				Format:   "jpeg",
				FileSize: 1024,
				Tags: []domain.TagInfo{
					{Type: "EXIF", Group: "Equipment", Name: "Make", Value: "Canon"},
					{Type: "EXIF", Group: "Equipment", Name: "Model", Value: "EOS 5D"},
				},
			},
			lang: "en",
		},
		{
			name: "with equipment in Russian",
			meta: domain.Metadata{
				Format:   "jpeg",
				FileSize: 1024,
				Tags: []domain.TagInfo{
					{Type: "EXIF", Group: "Equipment", Name: "Make", Value: "Canon"},
					{Type: "EXIF", Group: "Equipment", Name: "Model", Value: "EOS 5D"},
				},
			},
			lang: "ru",
		},
		{
			name: "with date and GPS",
			meta: domain.Metadata{
				Format: "jpeg",
				Tags: []domain.TagInfo{
					{Type: "EXIF", Group: "Location", Name: "GPSLatitude", Value: 55.7558},
					{Type: "EXIF", Group: "Location", Name: "GPSLongitude", Value: 37.6173},
				},
			},
			lang: "en",
		},
		{
			name: "with shooting params",
			meta: domain.Metadata{
				Format: "jpeg",
				Tags: []domain.TagInfo{
					{Type: "EXIF", Group: "Shooting", Name: "ISO", Value: 100},
					{Type: "EXIF", Group: "Shooting", Name: "FNumber", Value: 2.8},
					{Type: "EXIF", Group: "Shooting", Name: "FocalLength", Value: "50mm"},
					{Type: "EXIF", Group: "Shooting", Name: "ExposureTime", Value: "1/60"},
				},
			},
			lang: "en",
		},
		{
			name: "with photo properties",
			meta: domain.Metadata{
				Format: "jpeg",
				Tags: []domain.TagInfo{
					{Type: "EXIF", Group: "Photo", Name: "ImageWidth", Value: 1920},
					{Type: "EXIF", Group: "Photo", Name: "ImageHeight", Value: 1080},
					{Type: "EXIF", Group: "Photo", Name: "Orientation", Value: 1},
					{Type: "EXIF", Group: "Photo", Name: "ColorSpace", Value: "sRGB"},
				},
			},
			lang: "en",
		},
		{
			name: "with author",
			meta: domain.Metadata{
				Format: "jpeg",
				Tags: []domain.TagInfo{
					{Type: "EXIF", Group: "Author", Name: "Artist", Value: "John Doe"},
					{Type: "EXIF", Group: "Author", Name: "Copyright", Value: "2024"},
					{Type: "EXIF", Group: "Author", Name: "Creator", Value: "Jane"},
					{Type: "EXIF", Group: "Author", Name: "Credit", Value: "Test"},
					{Type: "EXIF", Group: "Author", Name: "Rights", Value: "CC0"},
				},
			},
			lang: "en",
		},
		{
			name: "with other properties",
			meta: domain.Metadata{
				Format: "jpeg",
				Tags: []domain.TagInfo{
					{Type: "EXIF", Group: "Other", Name: "CustomTag", Value: "CustomValue"},
					{Type: "EXIF", Group: "Other", Name: "NumberTag", Value: 42},
				},
			},
			lang: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := srv.localizeMetadata(tt.meta, tt.lang)
			if props == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}



func TestHandleAnalyze_WithLang(t *testing.T) {
	mockAnalyzer := fake.NewMockImageAnalyzer()
	mockAnalyzer.AnalyzeStreamFn = func(ctx context.Context, r io.Reader, name string, size int64) (*domain.ImageFile, error) {
		return &domain.ImageFile{
			Name: name,
			Metadata: domain.Metadata{
				Format:   "jpeg",
				FileSize: size,
			},
		}, nil
	}

	server := NewServer(mockAnalyzer)

	body, contentType := createMultipartRequest(t, "file", "test.jpg", []byte("fake"))
	req := httptest.NewRequest(http.MethodPost, "/analyze?lang=ru", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	server.handleAnalyze(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandleAnalyze_PartialSuccess(t *testing.T) {
	mockAnalyzer := fake.NewMockImageAnalyzer()
	mockAnalyzer.AnalyzeStreamFn = func(ctx context.Context, r io.Reader, name string, size int64) (*domain.ImageFile, error) {
		return &domain.ImageFile{
			Name: name,
			Metadata: domain.Metadata{
				Format:   "jpeg",
				FileSize: size,
				Tags: []domain.TagInfo{
					{Type: "EXIF", Group: "Equipment", Name: "Make", Value: "Nikon"},
					{Type: "EXIF", Group: "Equipment", Name: "Model", Value: "D850"},
				},
			},
		}, nil
	}

	server := NewServer(mockAnalyzer)

	body, contentType := createMultipartRequest(t, "file", "test.jpg", []byte("fake"))
	req := httptest.NewRequest(http.MethodPost, "/analyze", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	server.handleAnalyze(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp AnalyzeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Name != "test.jpg" {
		t.Errorf("expected name 'test.jpg', got %q", resp.Name)
	}

	if resp.Format != "jpeg" {
		t.Errorf("expected format 'jpeg', got %q", resp.Format)
	}
}

func TestHandleDemo_WrongMethod(t *testing.T) {
	mockAnalyzer := fake.NewMockImageAnalyzer()
	server := NewServer(mockAnalyzer)

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/demo", nil)
			rec := httptest.NewRecorder()

			server.handleDemo(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status %d for %s, got %d", http.StatusMethodNotAllowed, method, rec.Code)
			}
		})
	}
}

func TestAnalyzeResponse_JSONFormat(t *testing.T) {
	mockAnalyzer := fake.NewMockImageAnalyzer()
	mockAnalyzer.AnalyzeStreamFn = func(ctx context.Context, r io.Reader, name string, size int64) (*domain.ImageFile, error) {
		return &domain.ImageFile{
			Name: name,
			Metadata: domain.Metadata{
				Format:   "png",
				FileSize: size,
			},
		}, nil
	}

	server := NewServer(mockAnalyzer)

	body, contentType := createMultipartRequest(t, "file", "test.png", []byte("fake"))
	req := httptest.NewRequest(http.MethodPost, "/analyze", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	server.handleAnalyze(rec, req)

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", rec.Header().Get("Content-Type"))
	}

	// Verify the JSON can be parsed
	var resp AnalyzeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
}

func TestLoggingMiddleware_CapturesStatus(t *testing.T) {
	s := NewServer(fake.NewMockImageAnalyzer())

	mux := http.NewServeMux()
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	handler := s.loggingMiddleware(mux)

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func ExampleNewServer() {
	// This example demonstrates how to initialize the server.
	// We use a mock here, but in production, this would be the real service.
	mock := fake.NewMockImageAnalyzer()

	srv := NewServer(mock)

	// In a real application, you would run:
	// log.Fatal(srv.Start("8080"))

	if srv != nil {
		fmt.Println("Server created successfully")
	}

	// Output:
	// Server created successfully
}

// ============================================================================
// MIDDLEWARE TESTS
// ============================================================================

func TestMiddleware(t *testing.T) {
	s := NewServer(fake.NewMockImageAnalyzer())

	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	handler := wrap(mux,
		s.loggingMiddleware,
		s.rateLimitMiddleware(2, time.Second),
		s.timeoutMiddleware(100*time.Millisecond),
	)

	t.Run("Logging and OK", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", w.Code)
		}
	})

	t.Run("Rate Limit", func(t *testing.T) {
		// Reset rate limit maps for test
		rateLimitMu.Lock()
		lastSeen = make(map[string]time.Time)
		tokens = make(map[string]int)
		rateLimitMu.Unlock()

		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "1.2.3.4:1234"
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("Request %d expected OK, got %d", i, w.Code)
			}
		}

		// Third request should be rate limited
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status TooManyRequests, got %d", w.Code)
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		slowMux := http.NewServeMux()
		slowMux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
			w.Write([]byte("too slow"))
		})

		timeoutHandler := s.timeoutMiddleware(50 * time.Millisecond)(slowMux)

		req := httptest.NewRequest("GET", "/slow", nil)
		w := httptest.NewRecorder()
		timeoutHandler.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected status ServiceUnavailable (timeout), got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Request Timeout") {
			t.Errorf("Expected body to contain 'Request Timeout', got %q", w.Body.String())
		}
	})
}
