package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"asesamo-api/internal/adapter/http/dto"
	"asesamo-api/internal/adapter/storage"
	domainErrors "asesamo-api/internal/core/domain/errors"
	usecaseDoc "asesamo-api/internal/core/usecase/document"
	"asesamo-api/pkg/tenant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// DocumentHandler handles document upload, listing, review, and download endpoints.
type DocumentHandler struct {
	uploadUC      *usecaseDoc.UploadDocumentUseCase
	listUC        *usecaseDoc.ListMyDocumentsUseCase
	downloadUC    *usecaseDoc.DownloadDocumentUseCase
	listTypesUC   *usecaseDoc.ListDocumentTypesUseCase
	reviewUC      *usecaseDoc.ReviewDocumentUseCase
	listPendingUC *usecaseDoc.ListPendingDocumentsUseCase
}

// NewDocumentHandler constructs a new DocumentHandler.
func NewDocumentHandler(
	uploadUC *usecaseDoc.UploadDocumentUseCase,
	listUC *usecaseDoc.ListMyDocumentsUseCase,
	downloadUC *usecaseDoc.DownloadDocumentUseCase,
	listTypesUC *usecaseDoc.ListDocumentTypesUseCase,
) *DocumentHandler {
	return &DocumentHandler{
		uploadUC:    uploadUC,
		listUC:      listUC,
		downloadUC:  downloadUC,
		listTypesUC: listTypesUC,
	}
}

// SetReviewUseCases attaches review and pending list use cases.
func (h *DocumentHandler) SetReviewUseCases(reviewUC *usecaseDoc.ReviewDocumentUseCase, listPendingUC *usecaseDoc.ListPendingDocumentsUseCase) {
	h.reviewUC = reviewUC
	h.listPendingUC = listPendingUC
}

// SetListTypesUseCase attaches the document types catalog use case.
func (h *DocumentHandler) SetListTypesUseCase(uc *usecaseDoc.ListDocumentTypesUseCase) {
	h.listTypesUC = uc
}

// Upload handles POST /api/v1/documents/upload.
func (h *DocumentHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// 1. Verify caller authentication in context
	if _, err := tenant.FromContext(r.Context()); err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	// 2. Enforce 10 MB limit and parse multipart/form-data
	r.Body = http.MaxBytesReader(w, r.Body, storage.MaxFileSize)
	if err := r.ParseMultipartForm(storage.MaxFileSize); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) || strings.Contains(strings.ToLower(err.Error()), "request body too large") {
			dto.WriteError(w, domainErrors.ErrFileTooLarge)
			return
		}
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Error al procesar el formulario de carga: "+err.Error()))
		return
	}

	// 3. Extract uploaded file
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El archivo es obligatorio (campo 'file')."))
		return
	}
	defer file.Close()

	if fileHeader.Size > storage.MaxFileSize {
		dto.WriteError(w, domainErrors.ErrFileTooLarge)
		return
	}

	// 4. Inspect magic bytes for authentic PDF, JPEG, or PNG
	headerBuf := make([]byte, 512)
	n, readErr := io.ReadFull(file, headerBuf)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("No se pudo leer el archivo cargado."))
		return
	}
	headerBuf = headerBuf[:n]

	detectedMime, err := storage.ValidateMagicBytes(headerBuf)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	// Combine header bytes with the remainder of the file stream
	fullReader := io.MultiReader(bytes.NewReader(headerBuf), file)

	// 5. Extract and validate required metadata fields
	docTypeCode := strings.TrimSpace(r.FormValue("document_type_code"))
	if docTypeCode == "" {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El código del tipo de documento (document_type_code) es obligatorio."))
		return
	}

	issuedAtStr := strings.TrimSpace(r.FormValue("issued_at"))
	if issuedAtStr == "" {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("La fecha de emisión (issued_at) es obligatoria."))
		return
	}
	issuedAt, err := parseDate(issuedAtStr)
	if err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El formato de issued_at no es válido. Use YYYY-MM-DD."))
		return
	}

	var expiresAt *time.Time
	if expStr := strings.TrimSpace(r.FormValue("expires_at")); expStr != "" {
		expDate, err := parseDate(expStr)
		if err != nil {
			dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El formato de expires_at no es válido. Use YYYY-MM-DD."))
			return
		}
		expiresAt = expDate
	}

	// 6. Optional target identifiers
	var targetWorkerID *uuid.UUID
	if wIDStr := strings.TrimSpace(r.FormValue("worker_id")); wIDStr != "" {
		parsed, err := uuid.Parse(wIDStr)
		if err != nil {
			dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El formato de worker_id no es válido."))
			return
		}
		targetWorkerID = &parsed
	} else if wIDStr := strings.TrimSpace(r.FormValue("worker_user_id")); wIDStr != "" {
		parsed, err := uuid.Parse(wIDStr)
		if err != nil {
			dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El formato de worker_user_id no es válido."))
			return
		}
		targetWorkerID = &parsed
	}

	var targetContractorID *uuid.UUID
	if cIDStr := strings.TrimSpace(r.FormValue("contractor_id")); cIDStr != "" {
		parsed, err := uuid.Parse(cIDStr)
		if err != nil {
			dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El formato de contractor_id no es válido."))
			return
		}
		targetContractorID = &parsed
	}

	// 7. Invoke UploadDocumentUseCase
	doc, err := h.uploadUC.Execute(r.Context(), usecaseDoc.UploadDocumentInput{
		DocumentTypeCode: docTypeCode,
		FileName:         fileHeader.Filename,
		File:             fullReader,
		FileSize:         fileHeader.Size,
		IssuedAt:         issuedAt,
		ExpiresAt:        expiresAt,
		WorkerID:         targetWorkerID,
		ContractorID:     targetContractorID,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	// Ensure detected MIME is populated if not set
	if doc.MimeType == nil || *doc.MimeType == "" {
		doc.MimeType = &detectedMime
	}

	resp := dto.ToUploadDocumentResponse(doc)
	dto.WriteSuccess(w, http.StatusCreated, resp)
}

// ListMyDocuments handles GET /api/v1/documents/my-documents.
func (h *DocumentHandler) ListMyDocuments(w http.ResponseWriter, r *http.Request) {
	// 1. Verify caller authentication in context
	if _, err := tenant.FromContext(r.Context()); err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	// 2. Extract optional query filters
	query := r.URL.Query()
	var targetWorkerID *uuid.UUID
	if wIDStr := strings.TrimSpace(query.Get("worker_id")); wIDStr != "" {
		parsed, err := uuid.Parse(wIDStr)
		if err != nil {
			dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El formato de worker_id no es válido."))
			return
		}
		targetWorkerID = &parsed
	} else if wIDStr := strings.TrimSpace(query.Get("worker_user_id")); wIDStr != "" {
		parsed, err := uuid.Parse(wIDStr)
		if err != nil {
			dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El formato de worker_user_id no es válido."))
			return
		}
		targetWorkerID = &parsed
	}

	var targetContractorID *uuid.UUID
	if cIDStr := strings.TrimSpace(query.Get("contractor_id")); cIDStr != "" {
		parsed, err := uuid.Parse(cIDStr)
		if err != nil {
			dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El formato de contractor_id no es válido."))
			return
		}
		targetContractorID = &parsed
	}

	// 3. Invoke ListMyDocumentsUseCase
	docs, err := h.listUC.Execute(r.Context(), targetWorkerID, targetContractorID)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	resps := dto.ToMyDocumentItemResponses(docs)
	dto.WriteSuccess(w, http.StatusOK, resps)
}

// Download handles GET /api/v1/documents/{id}/download.
func (h *DocumentHandler) Download(w http.ResponseWriter, r *http.Request) {
	// 1. Verify caller authentication in context
	if _, err := tenant.FromContext(r.Context()); err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	// 2. Extract and parse document ID path parameter
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El ID del documento es obligatorio."))
		return
	}

	docID, err := uuid.Parse(idStr)
	if err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El formato del ID de documento no es válido."))
		return
	}

	// 3. Invoke DownloadDocumentUseCase
	stream, doc, err := h.downloadUC.Execute(r.Context(), docID)
	if err != nil {
		dto.WriteError(w, err)
		return
	}
	defer stream.Close()

	// 4. Set streaming response headers
	contentType := "application/octet-stream"
	if doc.MimeType != nil && *doc.MimeType != "" {
		contentType = *doc.MimeType
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", doc.FileName))
	if doc.FileSizeBytes != nil && *doc.FileSizeBytes > 0 {
		w.Header().Set("Content-Length", strconv.Itoa(*doc.FileSizeBytes))
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// 5. Stream content to client
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, stream)
}

// ListTypes returns the catalog of active document types.
// GET /api/v1/documents/types
func (h *DocumentHandler) ListTypes(w http.ResponseWriter, r *http.Request) {
	if h.listTypesUC == nil {
		dto.WriteSuccess(w, http.StatusOK, []dto.DocumentTypeResponse{})
		return
	}

	types, err := h.listTypesUC.Execute(r.Context())
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	res := make([]dto.DocumentTypeResponse, len(types))
	for i := range types {
		res[i] = dto.ToDocumentTypeResponse(&types[i])
	}

	dto.WriteSuccess(w, http.StatusOK, res)
}

// ListPending handles GET /api/v1/documents/pending.
func (h *DocumentHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	if h.listPendingUC == nil {
		dto.WriteSuccess(w, http.StatusOK, []dto.DocumentResponse{})
		return
	}

	statusFilter := r.URL.Query().Get("status")
	if statusFilter == "" {
		statusFilter = "pending_review"
	}

	docs, err := h.listPendingUC.Execute(r.Context(), statusFilter)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	res := make([]dto.DocumentResponse, len(docs))
	for i := range docs {
		res[i] = dto.ToDocumentResponse(&docs[i])
	}

	dto.WriteSuccess(w, http.StatusOK, res)
}

// Review handles POST /api/v1/documents/{id}/review.
func (h *DocumentHandler) Review(w http.ResponseWriter, r *http.Request) {
	if h.reviewUC == nil {
		dto.WriteError(w, domainErrors.ErrInternal)
		return
	}

	idStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(idStr)
	if err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("ID de documento no válido."))
		return
	}

	var req dto.ReviewDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Cuerpo de solicitud JSON no válido."))
		return
	}

	updatedDoc, err := h.reviewUC.Execute(r.Context(), usecaseDoc.ReviewDocumentInput{
		DocumentID:      docID,
		Action:          req.Action,
		RejectionReason: req.RejectionReason,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, dto.ToDocumentResponse(updatedDoc))
}

// parseDate parses strings formatted as YYYY-MM-DD or RFC3339.
func parseDate(val string) (*time.Time, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return nil, nil
	}
	for _, layout := range []string{"2006-01-02", time.RFC3339, "2006-01-02T15:04:05Z07:00"} {
		if t, err := time.Parse(layout, val); err == nil {
			utc := t.UTC()
			return &utc, nil
		}
	}
	return nil, domainErrors.ErrInvalidInput.WithMessage("Formato de fecha no válido. Use YYYY-MM-DD.")
}
