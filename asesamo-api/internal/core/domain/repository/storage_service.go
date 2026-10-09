package repository

import (
	"context"
	"io"
)

// StorageService defines abstract file storage operations for compliance documents.
type StorageService interface {
	Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	GetURL(ctx context.Context, key string) (string, error)
}
