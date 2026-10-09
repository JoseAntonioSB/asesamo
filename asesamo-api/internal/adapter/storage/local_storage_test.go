package storage_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"asesamo-api/internal/adapter/storage"
	domainErrors "asesamo-api/internal/core/domain/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Suite 1: TestValidateMagicBytes
func TestValidateMagicBytes(t *testing.T) {
	tests := []struct {
		name      string
		header    []byte
		expected  string
		expectErr bool
	}{
		{
			name:     "Valid PDF header",
			header:   []byte("%PDF-1.4\n%âãÏÓ\n"),
			expected: storage.ContentTypePDF,
		},
		{
			name:     "Valid PDF version 2.0",
			header:   []byte("%PDF-2.0\n"),
			expected: storage.ContentTypePDF,
		},
		{
			name:     "Valid JPEG standard JFIF",
			header:   []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
			expected: storage.ContentTypeJPEG,
		},
		{
			name:     "Valid JPEG EXIF",
			header:   []byte{0xFF, 0xD8, 0xFF, 0xE1, 0x00, 0x18, 0x45, 0x78, 0x69, 0x66},
			expected: storage.ContentTypeJPEG,
		},
		{
			name:     "Valid PNG signature",
			header:   []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D},
			expected: storage.ContentTypePNG,
		},
		{
			name:      "Empty header slice",
			header:    []byte{},
			expectErr: true,
		},
		{
			name:      "Nil header",
			header:    nil,
			expectErr: true,
		},
		{
			name:      "Short truncated JPEG header",
			header:    []byte{0xFF, 0xD8},
			expectErr: true,
		},
		{
			name:      "Spoofed executable MZ header",
			header:    []byte{'M', 'Z', 0x90, 0x00, 0x03, 0x00, 0x00, 0x00},
			expectErr: true,
		},
		{
			name:      "Spoofed Linux ELF binary",
			header:    []byte{0x7F, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00},
			expectErr: true,
		},
		{
			name:      "Plain text pretending to be document",
			header:    []byte("This is plain text and should be rejected."),
			expectErr: true,
		},
		{
			name:      "Shell script",
			header:    []byte("#!/bin/bash\nrm -rf /"),
			expectErr: true,
		},
		{
			name:      "HTML script injection",
			header:    []byte("<!DOCTYPE html><html><body><script>alert(1)</script></body></html>"),
			expectErr: true,
		},
		{
			name:      "Random binary noise",
			header:    []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.ValidateMagicBytes(tt.header)
			if tt.expectErr {
				assert.Error(t, err)
				assert.True(t, domainErrors.ErrInvalidFileType.Is(err) || err == domainErrors.ErrInvalidFileType)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

// Suite 2: TestLocalStorageService_Upload
func TestLocalStorageService_Upload(t *testing.T) {
	tempDir := t.TempDir()
	svc, err := storage.NewLocalStorageService(tempDir)
	require.NoError(t, err)
	ctx := context.Background()

	validPDF := append([]byte("%PDF-1.5\n"), []byte("sample pdf content for testing compliance")...)
	validJPEG := append([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}, []byte("sample jpeg data")...)
	validPNG := append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, []byte("sample png data")...)

	t.Run("Upload valid PDF", func(t *testing.T) {
		key := "tenant-1/worker/uuid-1/dc3.pdf"
		err := svc.Upload(ctx, key, bytes.NewReader(validPDF), int64(len(validPDF)), storage.ContentTypePDF)
		assert.NoError(t, err)

		fullPath := filepath.Join(tempDir, key)
		assert.FileExists(t, fullPath)
		data, readErr := os.ReadFile(fullPath)
		assert.NoError(t, readErr)
		assert.Equal(t, validPDF, data)
	})

	t.Run("Upload valid JPEG", func(t *testing.T) {
		key := "tenant-1/contractor/uuid-2/repse.jpg"
		err := svc.Upload(ctx, key, bytes.NewReader(validJPEG), int64(len(validJPEG)), storage.ContentTypeJPEG)
		assert.NoError(t, err)

		fullPath := filepath.Join(tempDir, key)
		assert.FileExists(t, fullPath)
	})

	t.Run("Upload valid PNG", func(t *testing.T) {
		key := "tenant-1/worker/uuid-3/ine.png"
		err := svc.Upload(ctx, key, bytes.NewReader(validPNG), int64(len(validPNG)), storage.ContentTypePNG)
		assert.NoError(t, err)

		fullPath := filepath.Join(tempDir, key)
		assert.FileExists(t, fullPath)
	})

	t.Run("Upload file size overflow via declared size", func(t *testing.T) {
		key := "tenant-1/worker/uuid-4/large.pdf"
		overflowSize := storage.MaxFileSize + 1
		err := svc.Upload(ctx, key, bytes.NewReader(validPDF), overflowSize, storage.ContentTypePDF)
		assert.ErrorIs(t, err, domainErrors.ErrFileTooLarge)
		assert.NoFileExists(t, filepath.Join(tempDir, key))
	})

	t.Run("Upload file size overflow via stream", func(t *testing.T) {
		key := "tenant-1/worker/uuid-5/stream_overflow.pdf"
		bigData := io.MultiReader(
			bytes.NewReader([]byte("%PDF-1.4\n")),
			bytes.NewReader(make([]byte, storage.MaxFileSize+1024)),
		)
		err := svc.Upload(ctx, key, bigData, 0, storage.ContentTypePDF)
		assert.ErrorIs(t, err, domainErrors.ErrFileTooLarge)
		assert.NoFileExists(t, filepath.Join(tempDir, key))
	})

	t.Run("Upload spoofed executable bytes", func(t *testing.T) {
		key := "tenant-1/worker/uuid-6/malicious.pdf"
		spoofedExe := []byte{'M', 'Z', 0x90, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04}
		err := svc.Upload(ctx, key, bytes.NewReader(spoofedExe), int64(len(spoofedExe)), storage.ContentTypePDF)
		assert.Error(t, err)
		assert.True(t, domainErrors.ErrInvalidFileType.Is(err))
		assert.NoFileExists(t, filepath.Join(tempDir, key))
	})

	t.Run("Upload mime mismatch", func(t *testing.T) {
		key := "tenant-1/worker/uuid-7/mismatch.pdf"
		// Send PNG bytes with PDF declared contentType
		err := svc.Upload(ctx, key, bytes.NewReader(validPNG), int64(len(validPNG)), storage.ContentTypePDF)
		assert.Error(t, err)
		assert.NoFileExists(t, filepath.Join(tempDir, key))
	})

	t.Run("Upload path traversal ../../etc/passwd", func(t *testing.T) {
		traversalKeys := []string{
			"../../etc/passwd",
			"/../../etc/shadow",
			"/etc/shadow",
			"tenant/../../../secret.txt",
			"..",
			".",
			"",
			"   ",
		}
		for _, tk := range traversalKeys {
			err := svc.Upload(ctx, tk, bytes.NewReader(validPDF), int64(len(validPDF)), storage.ContentTypePDF)
			assert.Error(t, err, "expected error for traversal key %q", tk)
			assert.True(t, domainErrors.ErrInvalidInput.Is(err), "expected ErrInvalidInput for %q", tk)
		}
	})

	t.Run("Upload overwrites existing file cleanly", func(t *testing.T) {
		key := "tenant-1/worker/overwrite.pdf"
		contentV1 := append([]byte("%PDF-1.4\n"), []byte("version 1")...)
		contentV2 := append([]byte("%PDF-1.4\n"), []byte("version 2 modified")...)

		err := svc.Upload(ctx, key, bytes.NewReader(contentV1), int64(len(contentV1)), storage.ContentTypePDF)
		require.NoError(t, err)

		err = svc.Upload(ctx, key, bytes.NewReader(contentV2), int64(len(contentV2)), storage.ContentTypePDF)
		require.NoError(t, err)

		readBack, err := os.ReadFile(filepath.Join(tempDir, key))
		require.NoError(t, err)
		assert.Equal(t, contentV2, readBack)
	})
}

// Suite 3: TestLocalStorageService_Download
func TestLocalStorageService_Download(t *testing.T) {
	tempDir := t.TempDir()
	svc, err := storage.NewLocalStorageService(tempDir)
	require.NoError(t, err)
	ctx := context.Background()

	validPDF := append([]byte("%PDF-1.4\n"), []byte("downloadable test pdf")...)
	key := "tenant-1/worker/doc.pdf"
	require.NoError(t, svc.Upload(ctx, key, bytes.NewReader(validPDF), int64(len(validPDF)), storage.ContentTypePDF))

	t.Run("Download valid file", func(t *testing.T) {
		rc, err := svc.Download(ctx, key)
		require.NoError(t, err)
		defer rc.Close()

		downloaded, err := io.ReadAll(rc)
		require.NoError(t, err)
		assert.Equal(t, validPDF, downloaded)
	})

	t.Run("Download non-existent file", func(t *testing.T) {
		rc, err := svc.Download(ctx, "tenant-1/worker/missing.pdf")
		assert.Error(t, err)
		assert.True(t, domainErrors.ErrDocumentNotFound.Is(err))
		assert.Nil(t, rc)
	})

	t.Run("Download path traversal rejection", func(t *testing.T) {
		rc, err := svc.Download(ctx, "../../etc/passwd")
		assert.Error(t, err)
		assert.True(t, domainErrors.ErrInvalidInput.Is(err))
		assert.Nil(t, rc)
	})

	t.Run("Download empty or dot key rejection", func(t *testing.T) {
		for _, k := range []string{"", ".", "   "} {
			rc, err := svc.Download(ctx, k)
			assert.Error(t, err)
			assert.True(t, domainErrors.ErrInvalidInput.Is(err))
			assert.Nil(t, rc)
		}
	})
}

// Suite 4: TestLocalStorageService_Delete
func TestLocalStorageService_Delete(t *testing.T) {
	tempDir := t.TempDir()
	svc, err := storage.NewLocalStorageService(tempDir)
	require.NoError(t, err)
	ctx := context.Background()

	validPDF := append([]byte("%PDF-1.4\n"), []byte("to be deleted")...)
	key := "tenant-1/worker/to_delete.pdf"
	require.NoError(t, svc.Upload(ctx, key, bytes.NewReader(validPDF), int64(len(validPDF)), storage.ContentTypePDF))

	t.Run("Delete existing file", func(t *testing.T) {
		err := svc.Delete(ctx, key)
		assert.NoError(t, err)
		assert.NoFileExists(t, filepath.Join(tempDir, key))
	})

	t.Run("Delete non-existent file is idempotent", func(t *testing.T) {
		err := svc.Delete(ctx, "tenant-1/worker/already_deleted.pdf")
		assert.NoError(t, err)
	})

	t.Run("Delete path traversal rejection", func(t *testing.T) {
		err := svc.Delete(ctx, "../../etc/passwd")
		assert.Error(t, err)
		assert.True(t, domainErrors.ErrInvalidInput.Is(err))
	})

	t.Run("Delete empty key does not delete root baseDir", func(t *testing.T) {
		err := svc.Delete(ctx, "")
		assert.Error(t, err)
		assert.True(t, domainErrors.ErrInvalidInput.Is(err))
		assert.DirExists(t, tempDir)
	})
}

// Suite 5: TestLocalStorageService_GetURL
func TestLocalStorageService_GetURL(t *testing.T) {
	tempDir := t.TempDir()
	svc, err := storage.NewLocalStorageService(tempDir)
	require.NoError(t, err)
	ctx := context.Background()

	url, err := svc.GetURL(ctx, "tenant-1/worker/uuid/doc.pdf")
	assert.NoError(t, err)
	assert.Equal(t, "/api/v1/documents/download?key=tenant-1/worker/uuid/doc.pdf", url)

	urlLeading, err := svc.GetURL(ctx, "/tenant-1/worker/uuid/doc.pdf")
	assert.NoError(t, err)
	assert.Equal(t, "/api/v1/documents/download?key=tenant-1/worker/uuid/doc.pdf", urlLeading)
}

// Suite 6: TestNewLocalStorageService
func TestNewLocalStorageService(t *testing.T) {
	t.Run("Creates base directory if not exists", func(t *testing.T) {
		nestedDir := filepath.Join(t.TempDir(), "sub", "uploads")
		svc, err := storage.NewLocalStorageService(nestedDir)
		assert.NoError(t, err)
		assert.NotNil(t, svc)
		assert.DirExists(t, nestedDir)
	})
}
