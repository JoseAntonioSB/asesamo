package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
)

// Allowed MIME types and max file size constant.
const (
	MaxFileSize     int64  = 10 * 1024 * 1024 // 10 MB
	ContentTypePDF  string = "application/pdf"
	ContentTypeJPEG string = "image/jpeg"
	ContentTypePNG  string = "image/png"
)

// ValidateMagicBytes inspects the initial byte header to verify authentic PDF, JPEG, or PNG files.
func ValidateMagicBytes(header []byte) (string, error) {
	if len(header) == 0 {
		return "", domainErrors.ErrInvalidFileType.WithMessage("El archivo está vacío o no contiene encabezado válido.")
	}

	// 1. PDF signature: starts with %PDF-
	if bytes.HasPrefix(header, []byte("%PDF-")) {
		return ContentTypePDF, nil
	}

	// 2. JPEG signature: starts with \xFF\xD8\xFF
	if len(header) >= 3 && header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
		return ContentTypeJPEG, nil
	}

	// 3. PNG signature: \x89PNG\r\n\x1a\n
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if bytes.HasPrefix(header, pngMagic) {
		return ContentTypePNG, nil
	}

	// Fallback to net/http sniffing
	sniffed := http.DetectContentType(header)
	if sniffed == ContentTypePDF || sniffed == ContentTypeJPEG || sniffed == ContentTypePNG {
		return sniffed, nil
	}

	return "", domainErrors.ErrInvalidFileType
}

type localStorageService struct {
	baseDir string
}

// NewLocalStorageService constructs a LocalStorageService that persists files to baseDir.
func NewLocalStorageService(baseDir string) (repository.StorageService, error) {
	cleanDir := filepath.Clean(baseDir)
	if err := os.MkdirAll(cleanDir, 0755); err != nil {
		return nil, fmt.Errorf("storage: failed to initialize storage directory %q: %w", cleanDir, err)
	}
	return &localStorageService{baseDir: cleanDir}, nil
}

func (s *localStorageService) resolvePath(key string) (string, error) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" || strings.ContainsRune(trimmed, 0) {
		return "", domainErrors.ErrInvalidInput.WithMessage("Clave de archivo no válida.")
	}

	if strings.HasPrefix(trimmed, "/") || filepath.IsAbs(trimmed) {
		return "", domainErrors.ErrInvalidInput.WithMessage("Clave de archivo no puede ser una ruta absoluta.")
	}

	cleanKey := filepath.Clean(trimmed)
	if cleanKey == "." || cleanKey == ".." || strings.HasPrefix(cleanKey, "..") || strings.Contains(cleanKey, "/../") {
		return "", domainErrors.ErrInvalidInput.WithMessage("Clave de archivo no válida.")
	}

	fullPath := filepath.Join(s.baseDir, cleanKey)
	rel, err := filepath.Rel(s.baseDir, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return "", domainErrors.ErrInvalidInput.WithMessage("Ruta de archivo fuera de los límites permitidos.")
	}

	return fullPath, nil
}

func (s *localStorageService) Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error {
	if size > MaxFileSize {
		return domainErrors.ErrFileTooLarge
	}

	fullPath, err := s.resolvePath(key)
	if err != nil {
		return err
	}

	// Peek first 512 bytes for magic byte verification
	headerBuf := make([]byte, 512)
	n, readErr := io.ReadFull(data, headerBuf)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		return fmt.Errorf("storage: failed to read file header: %w", readErr)
	}
	headerBuf = headerBuf[:n]

	detectedType, err := ValidateMagicBytes(headerBuf)
	if err != nil {
		return err
	}

	// If contentType was supplied, verify it matches detected type
	if contentType != "" && contentType != detectedType {
		return domainErrors.ErrInvalidFileType.WithMessage("El contenido del archivo no coincide con el tipo especificado.")
	}

	// Reconstruct stream using io.MultiReader and cap disk writing with io.LimitReader
	reconstructed := io.MultiReader(bytes.NewReader(headerBuf), data)
	limited := io.LimitReader(reconstructed, MaxFileSize+1)

	// Create parent directories
	parent := filepath.Dir(fullPath)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("storage: failed to create directory %q: %w", parent, err)
	}

	// Create target file with restricted permissions
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("storage: failed to create file %q: %w", fullPath, err)
	}

	written, copyErr := io.Copy(file, limited)
	closeErr := file.Close()

	if copyErr != nil {
		_ = os.Remove(fullPath)
		return fmt.Errorf("storage: failed to write file data: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(fullPath)
		return fmt.Errorf("storage: failed to close file %q: %w", fullPath, closeErr)
	}

	if written > MaxFileSize {
		_ = os.Remove(fullPath)
		return domainErrors.ErrFileTooLarge
	}

	return nil
}

func (s *localStorageService) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath, err := s.resolvePath(key)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domainErrors.ErrDocumentNotFound.WithMessage("El archivo no existe en el almacenamiento.")
		}
		return nil, fmt.Errorf("storage: failed to open file %q: %w", fullPath, err)
	}

	return file, nil
}

func (s *localStorageService) Delete(ctx context.Context, key string) error {
	fullPath, err := s.resolvePath(key)
	if err != nil {
		return err
	}

	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage: failed to delete file %q: %w", fullPath, err)
	}

	return nil
}

func (s *localStorageService) GetURL(ctx context.Context, key string) (string, error) {
	// For local storage, returns the relative API download path
	cleanKey := filepath.Clean(strings.TrimPrefix(key, "/"))
	return fmt.Sprintf("/api/v1/documents/download?key=%s", cleanKey), nil
}
