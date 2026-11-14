package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	domainErrors "asesamo-api/internal/core/domain/errors"
)

func TestR2StorageService_ConfigValidation(t *testing.T) {
	// Missing AccessKeyID
	_, err := NewR2StorageService(R2Config{
		SecretAccessKey: "secret",
		BucketName:      "test-bucket",
		AccountID:       "acc-123",
	})
	if err == nil {
		t.Errorf("expected error for missing AccessKeyID")
	}

	// Missing SecretAccessKey
	_, err = NewR2StorageService(R2Config{
		AccessKeyID: "key",
		BucketName:  "test-bucket",
		AccountID:   "acc-123",
	})
	if err == nil {
		t.Errorf("expected error for missing SecretAccessKey")
	}

	// Missing BucketName
	_, err = NewR2StorageService(R2Config{
		AccessKeyID:     "key",
		SecretAccessKey: "secret",
		AccountID:       "acc-123",
	})
	if err == nil {
		t.Errorf("expected error for missing BucketName")
	}

	// Valid config with account ID builds correct default endpoint
	svc, err := NewR2StorageService(R2Config{
		AccessKeyID:     "key",
		SecretAccessKey: "secret",
		BucketName:      "my-bucket",
		AccountID:       "acc-abc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r2Svc, ok := svc.(*r2StorageService)
	if !ok {
		t.Fatalf("expected *r2StorageService")
	}
	expectedEndpoint := "https://acc-abc.r2.cloudflarestorage.com"
	if r2Svc.endpoint != expectedEndpoint {
		t.Errorf("expected endpoint %q, got %q", expectedEndpoint, r2Svc.endpoint)
	}
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func TestR2StorageService_UploadDownloadDelete(t *testing.T) {
	var capturedAuthHeader string
	var capturedDateHeader string
	var capturedPayloadHash string
	var uploadedBytes []byte

	// In-memory mock transport (no socket binding required)
	mockClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) *http.Response {
			capturedAuthHeader = r.Header.Get("Authorization")
			capturedDateHeader = r.Header.Get("x-amz-date")
			capturedPayloadHash = r.Header.Get("x-amz-content-sha256")

			switch r.Method {
			case http.MethodPut:
				data, err := io.ReadAll(r.Body)
				if err != nil {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(bytes.NewReader(nil)),
						Header:     make(http.Header),
					}
				}
				uploadedBytes = data
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(nil)),
					Header:     make(http.Header),
				}

			case http.MethodGet:
				if strings.Contains(r.URL.Path, "missing.pdf") {
					return &http.Response{
						StatusCode: http.StatusNotFound,
						Body:       io.NopCloser(bytes.NewReader([]byte("NoSuchKey"))),
						Header:     make(http.Header),
					}
				}
				header := make(http.Header)
				header.Set("Content-Type", "application/pdf")
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte("%PDF-1.4 sample content from mock R2"))),
					Header:     header,
				}

			case http.MethodDelete:
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       io.NopCloser(bytes.NewReader(nil)),
					Header:     make(http.Header),
				}

			default:
				return &http.Response{
					StatusCode: http.StatusMethodNotAllowed,
					Body:       io.NopCloser(bytes.NewReader(nil)),
					Header:     make(http.Header),
				}
			}
		}),
	}

	svc, err := NewR2StorageService(R2Config{
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key-12345",
		BucketName:      "asesamo-test",
		Endpoint:        "https://r2.asesamo.local",
		HTTPClient:      mockClient,
		PublicURL:       "https://cdn.asesamo.com",
	})
	if err != nil {
		t.Fatalf("failed to init R2StorageService: %v", err)
	}

	ctx := context.Background()

	// 1. Upload valid PDF
	pdfPayload := []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\ntrailer\n<<>>\n%%EOF")
	err = svc.Upload(ctx, "contractors/c1/dc3.pdf", bytes.NewReader(pdfPayload), int64(len(pdfPayload)), "application/pdf")
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	if !strings.HasPrefix(capturedAuthHeader, "AWS4-HMAC-SHA256 Credential=test-access-key/") {
		t.Errorf("Authorization header malformed: %s", capturedAuthHeader)
	}
	if capturedDateHeader == "" {
		t.Errorf("missing x-amz-date header")
	}
	if capturedPayloadHash == "" {
		t.Errorf("missing x-amz-content-sha256 header")
	}
	if !bytes.Equal(uploadedBytes, pdfPayload) {
		t.Errorf("uploaded payload does not match original bytes")
	}

	// 2. Upload invalid/spoofed binary must fail before network
	spoofedPayload := []byte("MZ\x90\x00\x03\x00\x00\x00executable")
	err = svc.Upload(ctx, "contractors/c1/virus.exe", bytes.NewReader(spoofedPayload), int64(len(spoofedPayload)), "application/pdf")
	if err == nil {
		t.Errorf("expected magic byte error for spoofed binary")
	}

	// 3. Download existing file
	stream, err := svc.Download(ctx, "contractors/c1/dc3.pdf")
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	defer stream.Close()

	downloaded, err := io.ReadAll(stream)
	if err != nil {
		t.Fatalf("failed reading download stream: %v", err)
	}
	if !strings.Contains(string(downloaded), "%PDF-1.4") {
		t.Errorf("downloaded content invalid: %s", string(downloaded))
	}

	// 4. Download non-existent file returns ErrDocumentNotFound
	_, err = svc.Download(ctx, "contractors/c1/missing.pdf")
	if err == nil {
		t.Errorf("expected error downloading missing file")
	}
	if !errors.Is(err, domainErrors.ErrDocumentNotFound) {
		t.Errorf("expected ErrDocumentNotFound, got: %v", err)
	}

	// 5. Delete file
	err = svc.Delete(ctx, "contractors/c1/dc3.pdf")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 6. GetURL with public CDN URL configured
	urlStr, err := svc.GetURL(ctx, "contractors/c1/dc3.pdf")
	if err != nil {
		t.Fatalf("GetURL failed: %v", err)
	}
	expectedURL := "https://cdn.asesamo.com/contractors/c1/dc3.pdf"
	if urlStr != expectedURL {
		t.Errorf("expected URL %q, got %q", expectedURL, urlStr)
	}
}
