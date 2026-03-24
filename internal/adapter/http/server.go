package http

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/DementorAK/photometa/internal/platform/assets"
	"github.com/DementorAK/photometa/internal/domain"
	"github.com/DementorAK/photometa/internal/platform/locale"
	"github.com/DementorAK/photometa/internal/port"
)

// Server is the HTTP adapter for the ImageAnalyzer service.
type Server struct {
	service port.ImageAnalyzer
}

//go:embed demo.html
var demoHTML []byte

// NewServer creates a new HTTP server instance with the given service.
func NewServer(service port.ImageAnalyzer) *Server {
	return &Server{service: service}
}

// Start launches the HTTP server on the specified port.
// This method blocks until the server terminates.
func (s *Server) Start(port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/analyze", s.handleAnalyze)
	mux.HandleFunc("/locales", s.handleLocales)
	mux.HandleFunc("/demo", s.handleDemo)
	mux.HandleFunc("/icons/", s.handleIcons)

	addr := ":" + port
	
	// Apply middleware
	handler := wrap(mux,
		s.loggingMiddleware,
		s.rateLimitMiddleware(10, time.Minute), // 10 requests per minute per IP
		s.timeoutMiddleware(30*time.Second),
	)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Listen for shutdown signals in a separate goroutine
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Channel to receive server errors
	errChan := make(chan error, 1)

	go func() {
		slog.Info("Starting Web Server", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	// Block until a signal or a server error
	select {
	case err := <-errChan:
		return fmt.Errorf("server startup failed: %w", err)
	case sig := <-stopChan:
		slog.Info("Shutting down server...", "signal", sig)

		// Create a context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
		slog.Info("Server stopped")
	}

	return nil
}

type LocalizedProperty struct {
	Group    string `json:"group"`
	RawGroup string `json:"raw_group"` // For icon lookup
	Type     string `json:"type"`
	Value    any    `json:"value"`
	Synonym  string `json:"synonym"`
}

type AnalyzeResponse struct {
	Path     string              `json:"path"`
	Name     string              `json:"name"`
	Metadata []LocalizedProperty `json:"metadata"`
	Format   string              `json:"format"`
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size (e.g. 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Bad Request: failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Bad Request: missing 'file'", http.StatusBadRequest)
		return
	}
	defer file.Close()

	lang := r.FormValue("lang")
	if lang == "" {
		lang = "en"
	}

	img, err := s.service.AnalyzeStream(r.Context(), file, header.Filename, header.Size)
	if err != nil {
		slog.Error("Analysis failed", "filename", header.Filename, "error", err)
		http.Error(w, fmt.Sprintf("Analysis failed: %v", err), http.StatusInternalServerError)
		return
	}

	resp := AnalyzeResponse{
		Path:     img.Path,
		Name:     img.Name,
		Metadata: s.localizeMetadata(img.Metadata, lang),
		Format:   img.Metadata.Format, // Include format in response
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}

func (s *Server) localizeMetadata(m domain.Metadata, lang string) []LocalizedProperty {
	var props []LocalizedProperty

	add := func(tType, group, key string, val any) {
		if val == nil {
			return
		}
		// Convert value to string and localize if necessary/possible
		// For simplicity, we handle common types like in the GUI
		var sVal any = val
		switch v := val.(type) {
		case string:
			if v == "" || v == "<nil>" {
				return
			}
			sVal = locale.TForLocale(lang, v)
		case time.Time:
			if v.IsZero() {
				return
			}
			sVal = v.Format("02 Jan 2006 15:04:05")
		case *time.Time:
			if v == nil || v.IsZero() {
				return
			}
			sVal = v.Format("02 Jan 2006 15:04:05")
		case float64:
			if v == 0 {
				return
			}
		case int, int64:
			if v == 0 {
				// Don't skip if it's dimensions or something where 0 is valid? 
				// But GUI skips 0.
				return
			}
		}

		props = append(props, LocalizedProperty{
			Group:    locale.TForLocale(lang, group),
			RawGroup: group,
			Type:     tType,
			Value:    sVal,
			Synonym:  locale.TForLocale(lang, key),
		})
	}

	// File info
	add("File", "File", "Format", strings.ToUpper(m.Format))
	add("File", "File", "Size", fmt.Sprintf("%d %s", m.FileSize, locale.TForLocale(lang, "bytes")))

	var lat, lng *float64
	var latType, lngType string
	for _, t := range m.Tags {
		var val any = t.Value
		name := t.Name

		switch name {
		case "GPSLatitude":
			if v, ok := t.Value.(float64); ok {
				lat = &v
				latType = t.Type
			}
			continue
		case "GPSLongitude":
			if v, ok := t.Value.(float64); ok {
				lng = &v
				lngType = t.Type
			}
			continue
		case "FNumber":
			if f, ok := t.Value.(float64); ok && f > 0 {
				val = fmt.Sprintf("f/%.1f", f)
				name = "F-Number"
			}
		case "ExposureTime":
			name = "Exposure"
		case "FocalLength":
			name = "Focal Length"
		case "WhiteBalance":
			name = "White Balance"
		case "ColorSpace":
			name = "Color Space"
		case "DateTaken":
			name = "Date"
		}
		add(t.Type, t.Group, name, val)
	}

	if lat != nil && lng != nil {
		add(latType, "Location", "GPS", fmt.Sprintf("%.6f, %.6f", *lat, *lng))
	} else {
		if lat != nil {
			add(latType, "Location", "GPSLatitude", *lat)
		}
		if lng != nil {
			add(lngType, "Location", "GPSLongitude", *lng)
		}
	}

	return props
}

func (s *Server) handleLocales(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locale.GetLocales())
}

func (s *Server) handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(demoHTML)
}

func (s *Server) handleIcons(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/icons/")
	name = strings.TrimSuffix(name, ".svg")

	svg, err := assets.GetIcon(name)
	if err != nil {
		http.Error(w, "Icon not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write(svg)
}

type middleware func(http.Handler) http.Handler

func wrap(h http.Handler, mws ...middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Use a custom response writer to capture status code
		subWriter := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		
		next.ServeHTTP(subWriter, r)
		
		slog.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", subWriter.status,
			"duration", time.Since(start),
			"ip", r.RemoteAddr,
		)
	})
}

func (s *Server) timeoutMiddleware(timeout time.Duration) middleware {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, "Request Timeout")
	}
}

var (
	rateLimitMu sync.Mutex
	lastSeen   = make(map[string]time.Time)
	tokens     = make(map[string]int)
)

func (s *Server) rateLimitMiddleware(limit int, period time.Duration) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := strings.Split(r.RemoteAddr, ":")[0]

			rateLimitMu.Lock()
			now := time.Now()
			
			// Simple token bucket: reset tokens every period
			if now.Sub(lastSeen[ip]) > period {
				tokens[ip] = limit
				lastSeen[ip] = now
			}

			if tokens[ip] <= 0 {
				rateLimitMu.Unlock()
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			tokens[ip]--
			rateLimitMu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
