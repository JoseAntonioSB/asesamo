package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"asesamo-api/internal/adapter/http/dto"
	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/adapter/storage"
	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/document"
	"asesamo-api/pkg/tenant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// In-Memory Storage Stub for Handler Tests
type handlerMockStorage struct {
	files map[string][]byte
}

func newHandlerMockStorage() *handlerMockStorage {
	return &handlerMockStorage{files: make(map[string][]byte)}
}

func (s *handlerMockStorage) Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error {
	buf, err := io.ReadAll(data)
	if err != nil {
		return err
	}
	s.files[key] = buf
	return nil
}

func (s *handlerMockStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	data, ok := s.files[key]
	if !ok {
		return nil, domainErrors.ErrDocumentNotFound
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *handlerMockStorage) Delete(ctx context.Context, key string) error {
	delete(s.files, key)
	return nil
}

func (s *handlerMockStorage) GetURL(ctx context.Context, key string) (string, error) {
	return "/api/v1/documents/download?key=" + key, nil
}

// In-Memory Document Repository Stub for Handler Tests
type handlerMockDocRepo struct {
	docs        map[uuid.UUID]*entity.Document
	types       map[string]*entity.DocumentType
	contractors map[uuid.UUID]*entity.Contractor
	workerLinks map[uuid.UUID]uuid.UUID
}

func newHandlerMockDocRepo() *handlerMockDocRepo {
	dc3ID := uuid.New()
	repseID := uuid.New()
	ineID := uuid.New()

	return &handlerMockDocRepo{
		docs: make(map[uuid.UUID]*entity.Document),
		types: map[string]*entity.DocumentType{
			"DC3": {
				ID:                  dc3ID,
				Code:                "DC3",
				Name:                "Constancia DC-3",
				AppliesTo:           "worker",
				DefaultValidityDays: intPtr(365),
				IsActive:            true,
			},
			"REPSE": {
				ID:                  repseID,
				Code:                "REPSE",
				Name:                "Registro REPSE",
				AppliesTo:           "contractor",
				DefaultValidityDays: intPtr(365),
				IsActive:            true,
			},
			"INE": {
				ID:        ineID,
				Code:      "INE",
				Name:      "INE",
				AppliesTo: "worker",
				IsActive:  true,
			},
		},
		contractors: make(map[uuid.UUID]*entity.Contractor),
		workerLinks: make(map[uuid.UUID]uuid.UUID),
	}
}

func intPtr(i int) *int { return &i }

func (r *handlerMockDocRepo) Create(ctx context.Context, doc *entity.Document) error {
	r.docs[doc.ID] = doc
	return nil
}

func (r *handlerMockDocRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Document, error) {
	d, ok := r.docs[id]
	if !ok {
		return nil, domainErrors.ErrDocumentNotFound
	}
	return d, nil
}

func (r *handlerMockDocRepo) ListByWorker(ctx context.Context, workerUserID uuid.UUID) ([]entity.Document, error) {
	var list []entity.Document
	for _, d := range r.docs {
		if d.UserID != nil && *d.UserID == workerUserID {
			list = append(list, *d)
		}
	}
	return list, nil
}

func (r *handlerMockDocRepo) ListByContractor(ctx context.Context, contractorID uuid.UUID) ([]entity.Document, error) {
	var list []entity.Document
	for _, d := range r.docs {
		if d.ContractorID != nil && *d.ContractorID == contractorID {
			list = append(list, *d)
		}
	}
	return list, nil
}

func (r *handlerMockDocRepo) GetDocumentTypeByCode(ctx context.Context, code string) (*entity.DocumentType, error) {
	t, ok := r.types[code]
	if !ok {
		return nil, domainErrors.ErrDocumentTypeNotFound.WithMessage(fmt.Sprintf("Tipo de documento '%s' no válido.", code))
	}
	return t, nil
}

func (r *handlerMockDocRepo) ListRequiredDocumentTypes(ctx context.Context) ([]entity.DocumentType, error) {
	var list []entity.DocumentType
	for _, t := range r.types {
		list = append(list, *t)
	}
	return list, nil
}

func (r *handlerMockDocRepo) GetContractorByAdminUserID(ctx context.Context, adminUserID uuid.UUID) (*entity.Contractor, error) {
	for _, c := range r.contractors {
		if c.AdminUserID != nil && *c.AdminUserID == adminUserID {
			return c, nil
		}
	}
	return nil, domainErrors.ErrContractorNotFound
}

func (r *handlerMockDocRepo) GetContractorByWorkerUserID(ctx context.Context, workerUserID uuid.UUID) (*entity.Contractor, error) {
	cID, ok := r.workerLinks[workerUserID]
	if !ok {
		return nil, domainErrors.ErrContractorNotFound
	}
	c, ok := r.contractors[cID]
	if !ok {
		return nil, domainErrors.ErrContractorNotFound
	}
	return c, nil
}

func (r *handlerMockDocRepo) ListPendingReview(ctx context.Context, tenantID *uuid.UUID, statusFilter *entity.DocumentStatus) ([]entity.Document, error) {
	var list []entity.Document
	for _, d := range r.docs {
		if statusFilter != nil && d.Status != *statusFilter {
			continue
		}
		list = append(list, *d)
	}
	return list, nil
}

func (r *handlerMockDocRepo) UpdateReviewStatus(ctx context.Context, id uuid.UUID, status entity.DocumentStatus, reviewedBy uuid.UUID, rejectionReason *string) (*entity.Document, error) {
	d, ok := r.docs[id]
	if !ok {
		return nil, domainErrors.ErrDocumentNotFound
	}
	now := time.Now().UTC()
	d.Status = status
	d.ReviewedBy = &reviewedBy
	d.ReviewedAt = &now
	d.RejectionReason = rejectionReason
	d.UpdatedAt = now
	return d, nil
}

func createMultipartRequest(url string, fields map[string]string, fileFieldName string, fileName string, fileContent []byte) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			return nil, err
		}
	}

	if fileFieldName != "" && len(fileContent) > 0 {
		part, err := writer.CreateFormFile(fileFieldName, fileName)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(part, bytes.NewReader(fileContent)); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

// -------------------------------------------------------------
// Test 1: Upload Endpoint
// -------------------------------------------------------------
func TestDocumentHandler_Upload(t *testing.T) {
	storageMock := newHandlerMockStorage()
	repoMock := newHandlerMockDocRepo()

	uploadUC := document.NewUploadDocumentUseCase(repoMock, storageMock)
	listMyDocsUC := document.NewListMyDocumentsUseCase(repoMock)
	downloadUC := document.NewDownloadDocumentUseCase(repoMock, storageMock)
	listTypesUC := document.NewListDocumentTypesUseCase(repoMock)

	h := handler.NewDocumentHandler(uploadUC, listMyDocsUC, downloadUC, listTypesUC)

	tenantID := uuid.New()
	workerUserID := uuid.New()

	validPDFContent := append([]byte("%PDF-1.4\n"), []byte("sample pdf compliance content")...)
	validJPEGContent := append([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}, []byte("sample jpeg data")...)

	t.Run("successful PDF upload returns 201 Created and response envelope", func(t *testing.T) {
		fields := map[string]string{
			"document_type_code": "DC3",
			"issued_at":          "2026-01-15",
			"expires_at":         "2027-01-15",
		}

		req, err := createMultipartRequest("/api/v1/documents/upload", fields, "file", "dc3_cert.pdf", validPDFContent)
		require.NoError(t, err)

		ctx := tenant.WithTenant(req.Context(), tenantID)
		ctx = tenant.WithUserID(ctx, workerUserID)
		ctx = tenant.WithRole(ctx, "worker")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		h.Upload(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp struct {
			Data dto.UploadDocumentResponse `json:"data"`
		}
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.NotEqual(t, uuid.Nil, resp.Data.ID)
		assert.Equal(t, &workerUserID, resp.Data.UserID)
		assert.Nil(t, resp.Data.ContractorID)
		assert.Equal(t, "DC3", resp.Data.DocumentTypeCode)
		assert.Equal(t, "dc3_cert.pdf", resp.Data.FileName)
		assert.Equal(t, "pending_review", resp.Data.Status)
		assert.Equal(t, "application/pdf", *resp.Data.MimeType)
		assert.Equal(t, "2026-01-15", *resp.Data.IssuedAt)
		assert.Equal(t, "2027-01-15", *resp.Data.ExpiresAt)
	})

	t.Run("successful JPEG upload returns 201 Created", func(t *testing.T) {
		fields := map[string]string{
			"document_type_code": "INE",
			"issued_at":          "2025-06-01",
		}

		req, err := createMultipartRequest("/api/v1/documents/upload", fields, "file", "ine_photo.jpg", validJPEGContent)
		require.NoError(t, err)

		ctx := tenant.WithTenant(req.Context(), tenantID)
		ctx = tenant.WithUserID(ctx, workerUserID)
		ctx = tenant.WithRole(ctx, "worker")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		h.Upload(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("missing file returns 400 INVALID_INPUT", func(t *testing.T) {
		fields := map[string]string{
			"document_type_code": "DC3",
			"issued_at":          "2026-01-15",
		}

		req, err := createMultipartRequest("/api/v1/documents/upload", fields, "", "", nil)
		require.NoError(t, err)

		ctx := tenant.WithTenant(req.Context(), tenantID)
		ctx = tenant.WithUserID(ctx, workerUserID)
		ctx = tenant.WithRole(ctx, "worker")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		h.Upload(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var errResp dto.ErrorResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
		assert.Equal(t, "INVALID_INPUT", errResp.Error.Code)
	})

	t.Run("spoofed executable file returns 400 INVALID_FILE_TYPE", func(t *testing.T) {
		fields := map[string]string{
			"document_type_code": "DC3",
			"issued_at":          "2026-01-15",
		}

		fakeExe := []byte{'M', 'Z', 0x90, 0x00, 0x03, 0x00}
		req, err := createMultipartRequest("/api/v1/documents/upload", fields, "file", "malicious.pdf", fakeExe)
		require.NoError(t, err)

		ctx := tenant.WithTenant(req.Context(), tenantID)
		ctx = tenant.WithUserID(ctx, workerUserID)
		ctx = tenant.WithRole(ctx, "worker")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		h.Upload(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var errResp dto.ErrorResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
		assert.Equal(t, "INVALID_FILE_TYPE", errResp.Error.Code)
	})

	t.Run("missing document_type_code returns 400 INVALID_INPUT", func(t *testing.T) {
		fields := map[string]string{
			"issued_at": "2026-01-15",
		}

		req, err := createMultipartRequest("/api/v1/documents/upload", fields, "file", "doc.pdf", validPDFContent)
		require.NoError(t, err)

		ctx := tenant.WithTenant(req.Context(), tenantID)
		ctx = tenant.WithUserID(ctx, workerUserID)
		ctx = tenant.WithRole(ctx, "worker")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		h.Upload(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("missing authentication returns 401 UNAUTHORIZED", func(t *testing.T) {
		fields := map[string]string{
			"document_type_code": "DC3",
			"issued_at":          "2026-01-15",
		}

		req, err := createMultipartRequest("/api/v1/documents/upload", fields, "file", "doc.pdf", validPDFContent)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		h.Upload(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("file exceeding size limit returns 400 FILE_TOO_LARGE", func(t *testing.T) {
		fields := map[string]string{
			"document_type_code": "DC3",
			"issued_at":          "2026-01-15",
		}

		// Create a stream that exceeds MaxFileSize
		largeContent := append([]byte("%PDF-1.4\n"), make([]byte, storage.MaxFileSize+100)...)
		req, err := createMultipartRequest("/api/v1/documents/upload", fields, "file", "large.pdf", largeContent)
		require.NoError(t, err)

		ctx := tenant.WithTenant(req.Context(), tenantID)
		ctx = tenant.WithUserID(ctx, workerUserID)
		ctx = tenant.WithRole(ctx, "worker")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		h.Upload(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var errResp dto.ErrorResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
		assert.Equal(t, "FILE_TOO_LARGE", errResp.Error.Code)
	})
}

// -------------------------------------------------------------
// Test 2: List My Documents Endpoint
// -------------------------------------------------------------
func TestDocumentHandler_ListMyDocuments(t *testing.T) {
	storageMock := newHandlerMockStorage()
	repoMock := newHandlerMockDocRepo()

	uploadUC := document.NewUploadDocumentUseCase(repoMock, storageMock)
	listMyDocsUC := document.NewListMyDocumentsUseCase(repoMock)
	downloadUC := document.NewDownloadDocumentUseCase(repoMock, storageMock)
	listTypesUC := document.NewListDocumentTypesUseCase(repoMock)

	h := handler.NewDocumentHandler(uploadUC, listMyDocsUC, downloadUC, listTypesUC)

	tenantID := uuid.New()
	workerUserID := uuid.New()

	now := time.Now().UTC()
	docID := uuid.New()
	mime := "application/pdf"
	size := 1024

	repoMock.docs[docID] = &entity.Document{
		ID:               docID,
		UserID:           &workerUserID,
		DocumentTypeID:   repoMock.types["DC3"].ID,
		DocumentTypeCode: "DC3",
		DocumentTypeName: "Constancia DC-3",
		FileName:         "dc3.pdf",
		FileSizeBytes:    &size,
		MimeType:         &mime,
		Status:           entity.DocumentStatusPendingReview,
		IssuedAt:         &now,
		CreatedAt:        now,
	}

	t.Run("worker retrieves own documents wrapped in data envelope", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/my-documents", nil)
		ctx := tenant.WithTenant(req.Context(), tenantID)
		ctx = tenant.WithUserID(ctx, workerUserID)
		ctx = tenant.WithRole(ctx, "worker")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		h.ListMyDocuments(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Data []dto.MyDocumentItemResponse `json:"data"`
		}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		require.Len(t, resp.Data, 1)
		assert.Equal(t, docID, resp.Data[0].ID)
		assert.Equal(t, "DC3", resp.Data[0].DocumentTypeCode)
		assert.Equal(t, "dc3.pdf", resp.Data[0].FileName)
		assert.Equal(t, "pending_review", resp.Data[0].Status)
	})

	t.Run("empty list returns non-nil empty array in data envelope", func(t *testing.T) {
		otherWorkerID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/my-documents", nil)
		ctx := tenant.WithTenant(req.Context(), tenantID)
		ctx = tenant.WithUserID(ctx, otherWorkerID)
		ctx = tenant.WithRole(ctx, "worker")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		h.ListMyDocuments(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Data []dto.MyDocumentItemResponse `json:"data"`
		}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.NotNil(t, resp.Data)
		assert.Empty(t, resp.Data)
	})

	t.Run("missing authentication returns 401 UNAUTHORIZED", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/my-documents", nil)
		rec := httptest.NewRecorder()
		h.ListMyDocuments(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

// -------------------------------------------------------------
// Test 3: Download Endpoint
// -------------------------------------------------------------
func TestDocumentHandler_Download(t *testing.T) {
	storageMock := newHandlerMockStorage()
	repoMock := newHandlerMockDocRepo()

	uploadUC := document.NewUploadDocumentUseCase(repoMock, storageMock)
	listMyDocsUC := document.NewListMyDocumentsUseCase(repoMock)
	downloadUC := document.NewDownloadDocumentUseCase(repoMock, storageMock)
	listTypesUC := document.NewListDocumentTypesUseCase(repoMock)

	h := handler.NewDocumentHandler(uploadUC, listMyDocsUC, downloadUC, listTypesUC)

	tenantID := uuid.New()
	workerUserID := uuid.New()

	docID := uuid.New()
	r2Key := "tenant-1/workers/user-1/file.pdf"
	pdfContent := []byte("%PDF-1.4\ncontent to download")
	mime := "application/pdf"
	size := len(pdfContent)

	storageMock.files[r2Key] = pdfContent
	repoMock.docs[docID] = &entity.Document{
		ID:            docID,
		UserID:        &workerUserID,
		R2Key:         r2Key,
		FileName:      "download_me.pdf",
		FileSizeBytes: &size,
		MimeType:      &mime,
	}

	// Router for Chi path param extraction
	r := chi.NewRouter()
	r.Get("/api/v1/documents/{id}/download", h.Download)

	t.Run("successful download streams bytes with correct headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/"+docID.String()+"/download", nil)
		ctx := tenant.WithTenant(req.Context(), tenantID)
		ctx = tenant.WithUserID(ctx, workerUserID)
		ctx = tenant.WithRole(ctx, "worker")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/pdf", rec.Header().Get("Content-Type"))
		assert.Equal(t, `inline; filename="download_me.pdf"`, rec.Header().Get("Content-Disposition"))
		assert.Equal(t, fmt.Sprintf("%d", size), rec.Header().Get("Content-Length"))
		assert.Equal(t, pdfContent, rec.Body.Bytes())
	})

	t.Run("non-existent document returns 404 DOCUMENT_NOT_FOUND", func(t *testing.T) {
		missingID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/"+missingID.String()+"/download", nil)
		ctx := tenant.WithTenant(req.Context(), tenantID)
		ctx = tenant.WithUserID(ctx, workerUserID)
		ctx = tenant.WithRole(ctx, "worker")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("unauthorized worker trying to download another worker's file returns 403 FORBIDDEN", func(t *testing.T) {
		unauthorizedWorkerID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/"+docID.String()+"/download", nil)
		ctx := tenant.WithTenant(req.Context(), tenantID)
		ctx = tenant.WithUserID(ctx, unauthorizedWorkerID)
		ctx = tenant.WithRole(ctx, "worker")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("invalid UUID in URL param returns 400 INVALID_INPUT", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/not-a-uuid/download", nil)
		ctx := tenant.WithTenant(req.Context(), tenantID)
		ctx = tenant.WithUserID(ctx, workerUserID)
		ctx = tenant.WithRole(ctx, "worker")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("missing authentication returns 401 UNAUTHORIZED", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/"+docID.String()+"/download", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

// -------------------------------------------------------------
// Test 4: ListTypes Endpoint
// -------------------------------------------------------------
func TestDocumentHandler_ListTypes(t *testing.T) {
	storageMock := newHandlerMockStorage()
	repoMock := newHandlerMockDocRepo()

	uploadUC := document.NewUploadDocumentUseCase(repoMock, storageMock)
	listMyDocsUC := document.NewListMyDocumentsUseCase(repoMock)
	downloadUC := document.NewDownloadDocumentUseCase(repoMock, storageMock)
	listTypesUC := document.NewListDocumentTypesUseCase(repoMock)

	h := handler.NewDocumentHandler(uploadUC, listMyDocsUC, downloadUC, listTypesUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/types", nil)
	rec := httptest.NewRecorder()

	h.ListTypes(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data []dto.DocumentTypeResponse `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Len(t, resp.Data, 3)
}

// -------------------------------------------------------------
// Test 5: Review Document Endpoint (Approve / Reject)
// -------------------------------------------------------------
func TestDocumentHandler_Review(t *testing.T) {
	storageMock := newHandlerMockStorage()
	repoMock := newHandlerMockDocRepo()

	uploadUC := document.NewUploadDocumentUseCase(repoMock, storageMock)
	listMyDocsUC := document.NewListMyDocumentsUseCase(repoMock)
	downloadUC := document.NewDownloadDocumentUseCase(repoMock, storageMock)
	listTypesUC := document.NewListDocumentTypesUseCase(repoMock)
	reviewUC := document.NewReviewDocumentUseCase(repoMock)
	listPendingUC := document.NewListPendingDocumentsUseCase(repoMock)

	h := handler.NewDocumentHandler(uploadUC, listMyDocsUC, downloadUC, listTypesUC)
	h.SetReviewUseCases(reviewUC, listPendingUC)

	ehsID := uuid.New()
	workerID := uuid.New()
	tenantID := uuid.New()

	docID := uuid.New()
	repoMock.docs[docID] = &entity.Document{
		ID:               docID,
		UserID:           &workerID,
		DocumentTypeCode: "DC3",
		FileName:         "capacitacion.pdf",
		Status:           entity.DocumentStatusPendingReview,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	t.Run("EHS Manager successfully approves pending document", func(t *testing.T) {
		r := chi.NewRouter()
		r.Post("/api/v1/documents/{id}/review", h.Review)

		body := bytes.NewBufferString(`{"action":"approve"}`)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/documents/%s/review", docID), body)
		req.Header.Set("Content-Type", "application/json")

		ctx := req.Context()
		ctx = tenant.WithTenant(ctx, tenantID)
		ctx = tenant.WithUserID(ctx, ehsID)
		ctx = tenant.WithRole(ctx, "ehs_manager")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Data dto.DocumentResponse `json:"data"`
		}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, entity.DocumentStatusApproved, resp.Data.Status)
		assert.Equal(t, docID, resp.Data.ID)
	})

	t.Run("EHS Manager successfully rejects document with reason", func(t *testing.T) {
		r := chi.NewRouter()
		r.Post("/api/v1/documents/{id}/review", h.Review)

		reason := "Documento ilegible en el reverso"
		body := bytes.NewBufferString(fmt.Sprintf(`{"action":"reject","rejection_reason":%q}`, reason))
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/documents/%s/review", docID), body)
		req.Header.Set("Content-Type", "application/json")

		ctx := req.Context()
		ctx = tenant.WithTenant(ctx, tenantID)
		ctx = tenant.WithUserID(ctx, ehsID)
		ctx = tenant.WithRole(ctx, "plant_admin")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Data dto.DocumentResponse `json:"data"`
		}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, entity.DocumentStatusRejected, resp.Data.Status)
		assert.NotNil(t, resp.Data.RejectionReason)
		assert.Equal(t, reason, *resp.Data.RejectionReason)
	})

	t.Run("Rejecting without reason returns 400 Bad Request", func(t *testing.T) {
		r := chi.NewRouter()
		r.Post("/api/v1/documents/{id}/review", h.Review)

		body := bytes.NewBufferString(`{"action":"reject","rejection_reason":""}`)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/documents/%s/review", docID), body)
		req.Header.Set("Content-Type", "application/json")

		ctx := req.Context()
		ctx = tenant.WithTenant(ctx, tenantID)
		ctx = tenant.WithUserID(ctx, ehsID)
		ctx = tenant.WithRole(ctx, "ehs_manager")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Worker role cannot review documents (403 Forbidden)", func(t *testing.T) {
		r := chi.NewRouter()
		r.Post("/api/v1/documents/{id}/review", h.Review)

		body := bytes.NewBufferString(`{"action":"approve"}`)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/documents/%s/review", docID), body)
		req.Header.Set("Content-Type", "application/json")

		ctx := req.Context()
		ctx = tenant.WithTenant(ctx, tenantID)
		ctx = tenant.WithUserID(ctx, workerID)
		ctx = tenant.WithRole(ctx, "worker")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

// -------------------------------------------------------------
// Test 6: ListPending Endpoint
// -------------------------------------------------------------
func TestDocumentHandler_ListPending(t *testing.T) {
	storageMock := newHandlerMockStorage()
	repoMock := newHandlerMockDocRepo()

	uploadUC := document.NewUploadDocumentUseCase(repoMock, storageMock)
	listMyDocsUC := document.NewListMyDocumentsUseCase(repoMock)
	downloadUC := document.NewDownloadDocumentUseCase(repoMock, storageMock)
	listTypesUC := document.NewListDocumentTypesUseCase(repoMock)
	reviewUC := document.NewReviewDocumentUseCase(repoMock)
	listPendingUC := document.NewListPendingDocumentsUseCase(repoMock)

	h := handler.NewDocumentHandler(uploadUC, listMyDocsUC, downloadUC, listTypesUC)
	h.SetReviewUseCases(reviewUC, listPendingUC)

	ehsID := uuid.New()
	workerID := uuid.New()
	tenantID := uuid.New()

	doc1 := uuid.New()
	doc2 := uuid.New()
	repoMock.docs[doc1] = &entity.Document{
		ID:               doc1,
		UserID:           &workerID,
		DocumentTypeCode: "DC3",
		FileName:         "dc3.pdf",
		Status:           entity.DocumentStatusPendingReview,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	repoMock.docs[doc2] = &entity.Document{
		ID:               doc2,
		UserID:           &workerID,
		DocumentTypeCode: "INE",
		FileName:         "ine.pdf",
		Status:           entity.DocumentStatusApproved,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	r := chi.NewRouter()
	r.Get("/api/v1/documents/pending", h.ListPending)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/pending?status=pending_review", nil)
	ctx := req.Context()
	ctx = tenant.WithTenant(ctx, tenantID)
	ctx = tenant.WithUserID(ctx, ehsID)
	ctx = tenant.WithRole(ctx, "ehs_manager")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data []dto.DocumentResponse `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Len(t, resp.Data, 1)
	assert.Equal(t, doc1, resp.Data[0].ID)
}
