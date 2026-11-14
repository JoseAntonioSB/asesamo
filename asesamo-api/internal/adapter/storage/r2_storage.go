package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
)

// R2Config encapsulates connection parameters for Cloudflare R2 S3-compatible storage.
type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Endpoint        string // Optional: overrides default https://<AccountID>.r2.cloudflarestorage.com
	PublicURL       string // Optional: public CDN or r2.dev domain
	Region          string // Optional: defaults to "auto"
	HTTPClient      *http.Client
}

type r2StorageService struct {
	accountID       string
	accessKeyID     string
	secretAccessKey string
	bucketName      string
	endpoint        string
	publicURL       string
	region          string
	client          *http.Client
}

// NewR2StorageService constructs an R2-backed repository.StorageService implementation.
func NewR2StorageService(cfg R2Config) (repository.StorageService, error) {
	if strings.TrimSpace(cfg.AccessKeyID) == "" {
		return nil, domainErrors.ErrInvalidInput.WithMessage("R2_ACCESS_KEY_ID es obligatorio.")
	}
	if strings.TrimSpace(cfg.SecretAccessKey) == "" {
		return nil, domainErrors.ErrInvalidInput.WithMessage("R2_SECRET_ACCESS_KEY es obligatorio.")
	}
	if strings.TrimSpace(cfg.BucketName) == "" {
		return nil, domainErrors.ErrInvalidInput.WithMessage("R2_BUCKET_NAME es obligatorio.")
	}

	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		if strings.TrimSpace(cfg.AccountID) == "" {
			return nil, domainErrors.ErrInvalidInput.WithMessage("R2_ACCOUNT_ID o R2_ENDPOINT es obligatorio.")
		}
		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", strings.TrimSpace(cfg.AccountID))
	}
	// Strip trailing slashes
	endpoint = strings.TrimRight(endpoint, "/")

	region := strings.TrimSpace(cfg.Region)
	if region == "" {
		region = "auto"
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &r2StorageService{
		accountID:       strings.TrimSpace(cfg.AccountID),
		accessKeyID:     strings.TrimSpace(cfg.AccessKeyID),
		secretAccessKey: strings.TrimSpace(cfg.SecretAccessKey),
		bucketName:      strings.TrimSpace(cfg.BucketName),
		endpoint:        endpoint,
		publicURL:       strings.TrimRight(strings.TrimSpace(cfg.PublicURL), "/"),
		region:          region,
		client:          httpClient,
	}, nil
}

func (s *r2StorageService) cleanKey(key string) string {
	clean := path.Clean(strings.TrimSpace(key))
	clean = strings.TrimPrefix(clean, "/")
	return clean
}

func (s *r2StorageService) objectURL(key string) string {
	cleanKey := s.cleanKey(key)
	return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucketName, cleanKey)
}

// Upload stores a file stream to Cloudflare R2 bucket after magic byte validation.
func (s *r2StorageService) Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error {
	if size > MaxFileSize {
		return domainErrors.ErrFileTooLarge
	}

	cleanKey := s.cleanKey(key)
	if cleanKey == "" || cleanKey == "." || strings.HasPrefix(cleanKey, "..") {
		return domainErrors.ErrInvalidInput.WithMessage("Clave de archivo no válida.")
	}

	// 1. Inspect magic bytes for Authentic PDF, JPEG, or PNG
	headerBuf := make([]byte, 512)
	n, readErr := io.ReadFull(data, headerBuf)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		return fmt.Errorf("r2: failed to read file header: %w", readErr)
	}
	headerBuf = headerBuf[:n]

	detectedType, err := ValidateMagicBytes(headerBuf)
	if err != nil {
		return err
	}

	if contentType != "" && contentType != detectedType {
		return domainErrors.ErrInvalidFileType.WithMessage("El contenido del archivo no coincide con el tipo especificado.")
	}
	finalContentType := detectedType

	// Reconstruct stream and read all payload bytes (bounded to 10MB) to compute SHA256
	fullPayload := io.MultiReader(bytes.NewReader(headerBuf), data)
	payloadBytes, err := io.ReadAll(io.LimitReader(fullPayload, MaxFileSize+1))
	if err != nil {
		return fmt.Errorf("r2: failed to read payload: %w", err)
	}

	if int64(len(payloadBytes)) > MaxFileSize {
		return domainErrors.ErrFileTooLarge
	}

	reqURL := s.objectURL(cleanKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("r2: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", finalContentType)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(payloadBytes)))

	// Sign request with AWS SigV4
	if err := s.signRequest(req, payloadBytes); err != nil {
		return fmt.Errorf("r2: failed to sign request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("r2: upload HTTP error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("r2: upload failed with HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// Download retrieves a readable stream for the specified key from Cloudflare R2.
func (s *r2StorageService) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	cleanKey := s.cleanKey(key)
	if cleanKey == "" || cleanKey == "." || strings.HasPrefix(cleanKey, "..") {
		return nil, domainErrors.ErrInvalidInput.WithMessage("Clave de archivo no válida.")
	}

	reqURL := s.objectURL(cleanKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("r2: failed to create download request: %w", err)
	}

	if err := s.signRequest(req, nil); err != nil {
		return nil, fmt.Errorf("r2: failed to sign download request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("r2: download HTTP error: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, domainErrors.ErrDocumentNotFound.WithMessage("El archivo no existe en Cloudflare R2.")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		resp.Body.Close()
		return nil, fmt.Errorf("r2: download failed with HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return resp.Body, nil
}

// Delete permanently removes an object from the Cloudflare R2 bucket.
func (s *r2StorageService) Delete(ctx context.Context, key string) error {
	cleanKey := s.cleanKey(key)
	if cleanKey == "" || cleanKey == "." || strings.HasPrefix(cleanKey, "..") {
		return domainErrors.ErrInvalidInput.WithMessage("Clave de archivo no válida.")
	}

	reqURL := s.objectURL(cleanKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return fmt.Errorf("r2: failed to create delete request: %w", err)
	}

	if err := s.signRequest(req, nil); err != nil {
		return fmt.Errorf("r2: failed to sign delete request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("r2: delete HTTP error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("r2: delete failed with HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// GetURL returns the download URL for the object.
func (s *r2StorageService) GetURL(ctx context.Context, key string) (string, error) {
	cleanKey := s.cleanKey(key)
	if s.publicURL != "" {
		return fmt.Sprintf("%s/%s", s.publicURL, cleanKey), nil
	}
	// Fallback to internal authenticated backend streaming endpoint
	return fmt.Sprintf("/api/v1/documents/download?key=%s", url.QueryEscape(cleanKey)), nil
}

// -----------------------------------------------------------------------------
// AWS Signature Version 4 (SigV4) Implementation for S3/R2
// -----------------------------------------------------------------------------

func (s *r2StorageService) signRequest(req *http.Request, payload []byte) error {
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	// Host header
	host := req.URL.Host
	req.Header.Set("Host", host)
	req.Header.Set("x-amz-date", amzDate)

	// Payload SHA256 hash
	var payloadHash string
	if len(payload) > 0 {
		hash := sha256.Sum256(payload)
		payloadHash = hex.EncodeToString(hash[:])
	} else {
		hash := sha256.Sum256([]byte(""))
		payloadHash = hex.EncodeToString(hash[:])
	}
	req.Header.Set("x-amz-content-sha256", payloadHash)

	// Canonical headers
	signedHeaderKeys := []string{"host", "x-amz-content-sha256", "x-amz-date"}
	if req.Header.Get("Content-Type") != "" {
		signedHeaderKeys = append(signedHeaderKeys, "content-type")
	}
	sort.Strings(signedHeaderKeys)

	var canonicalHeaders strings.Builder
	for _, k := range signedHeaderKeys {
		val := strings.TrimSpace(req.Header.Get(k))
		canonicalHeaders.WriteString(fmt.Sprintf("%s:%s\n", strings.ToLower(k), val))
	}
	signedHeadersStr := strings.Join(signedHeaderKeys, ";")

	// Canonical URI
	canonicalURI := req.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	// Canonical Query String
	canonicalQuery := req.URL.Query().Encode()

	// Canonical Request
	canonicalRequest := fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n%s",
		req.Method,
		canonicalURI,
		canonicalQuery,
		canonicalHeaders.String(),
		signedHeadersStr,
		payloadHash,
	)

	reqHash := sha256.Sum256([]byte(canonicalRequest))
	hashedCanonicalRequest := hex.EncodeToString(reqHash[:])

	// Credential scope
	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, s.region)

	// String to sign
	stringToSign := fmt.Sprintf(
		"AWS4-HMAC-SHA256\n%s\n%s\n%s",
		amzDate,
		credentialScope,
		hashedCanonicalRequest,
	)

	// Signing key computation
	kDate := hmacSHA256([]byte("AWS4"+s.secretAccessKey), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(s.region))
	kService := hmacSHA256(kRegion, []byte("s3"))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))

	// Signature
	signature := hex.EncodeToString(hmacSHA256(kSigning, []byte(stringToSign)))

	// Authorization header
	authHeader := fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.accessKeyID,
		credentialScope,
		signedHeadersStr,
		signature,
	)
	req.Header.Set("Authorization", authHeader)

	return nil
}

func hmacSHA256(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
