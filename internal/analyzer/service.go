package analyzer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/DementorAK/photometa/internal/format/imageformat"
	"github.com/DementorAK/photometa/internal/analyzer/filler"
	"github.com/DementorAK/photometa/internal/domain"
	"github.com/DementorAK/photometa/internal/port"
)

// Compile-time check: Service must implement port.ImageAnalyzer
var _ port.ImageAnalyzer = (*Service)(nil)

// Service is the core application service that implements port.ImageAnalyzer.
type Service struct {
	logger port.Logger
}

// NewService creates a new Service with the given dependencies.
func NewService(logger port.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

// AnalyzeStream processes an image from a generic reader (stdin, http request).
func (s *Service) AnalyzeStream(_ context.Context, r io.Reader, name string, size int64) (*domain.ImageFile, error) {
	s.logger.Info("analyzing stream", "name", name, "size", size)
	buf, err := io.ReadAll(r)
	if err != nil {
		s.logger.Error("failed to read stream", "name", name, "error", err)
		return nil, fmt.Errorf("failed to read stream: %w", err)
	}

	if size == 0 {
		size = int64(len(buf))
	}

	meta := domain.Metadata{
		FileSize: size,
		Format:   string(imageformat.Detect(buf)),
	}

	filler.Fill(bytes.NewReader(buf), &meta, s.logger)

	s.logger.Debug("stream analysis completed", "name", name, "format", meta.Format)

	return &domain.ImageFile{
		Path:     "",
		Name:     name,
		Metadata: meta,
	}, nil
}

// AnalyzeFile processes a single file path.
func (s *Service) AnalyzeFile(_ context.Context, path string) (*domain.ImageFile, error) {
	s.logger.Info("analyzing file", "path", path)
	info, err := os.Stat(path)
	if err != nil {
		s.logger.Error("failed to stat file", "path", path, "error", err)
		return nil, err
	}

	if info.IsDir() {
		err := fmt.Errorf("path is a directory")
		s.logger.Error("cannot analyze directory as file", "path", path)
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		s.logger.Error("failed to read file", "path", path, "error", err)
		return nil, err
	}

	meta := domain.Metadata{
		FileSize: info.Size(),
		Format:   string(imageformat.Detect(data)),
	}

	// Fill file times
	meta.Tags = append(meta.Tags, domain.TagInfo{
		Type:  "File",
		Group: "Photo",
		Name:  "FileModifyTime",
		Value: info.ModTime(),
	})

	filler.Fill(bytes.NewReader(data), &meta, s.logger)

	s.logger.Debug("file analysis completed", "path", path, "format", meta.Format)

	return &domain.ImageFile{
		Path:     path,
		Name:     filepath.Base(path),
		Metadata: meta,
	}, nil
}

// ScanDirectory walks a directory and returns all found images.
func (s *Service) ScanDirectory(ctx context.Context, root string) ([]domain.ImageFile, error) {
	s.logger.Info("scanning directory", "root", root)
	var results []domain.ImageFile
	var errorCount int

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.logger.Warn("failed to access path during scan", "path", path, "error", err)
			errorCount++
			return nil
		}

		// Check context for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".tif" && ext != ".tiff" && ext != ".png" && ext != ".webp" && ext != ".gif" {
			return nil
		}

		img, err := s.AnalyzeFile(ctx, path)
		if err == nil {
			results = append(results, *img)
		} else {
			s.logger.Warn("failed to analyze file", "path", path, "error", err)
			errorCount++
		}
		return nil
	})

	if errorCount > 0 {
		s.logger.Warn("directory scan completed with errors", "root", root, "error_count", errorCount, "success_count", len(results))
	} else {
		s.logger.Info("directory scan completed successfully", "root", root, "success_count", len(results))
	}

	return results, err
}
