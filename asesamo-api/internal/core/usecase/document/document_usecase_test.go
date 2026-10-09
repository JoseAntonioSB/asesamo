package document_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/document"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

// Mock Storage
type mockStorageService struct {
	uploads map[string][]byte
}

func newMockStorage() *mockStorageService {
	return &mockStorageService{uploads: make(map[string][]byte)}
}

func (m *mockStorageService) Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error {
	buf, err := io.ReadAll(data)
	if err != nil {
		return err
	}
	m.uploads[key] = buf
	return nil
}

func (m *mockStorageService) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	data, ok := m.uploads[key]
	if !ok {
		return nil, domainErrors.ErrDocumentNotFound
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *mockStorageService) Delete(ctx context.Context, key string) error {
	delete(m.uploads, key)
	return nil
}

func (m *mockStorageService) GetURL(ctx context.Context, key string) (string, error) {
	return "/api/v1/documents/download?key=" + key, nil
}

// Mock DocumentRepository
type mockDocumentRepository struct {
	docs        map[uuid.UUID]*entity.Document
	types       map[string]*entity.DocumentType
	contractors map[uuid.UUID]*entity.Contractor
	workerLinks map[uuid.UUID]uuid.UUID // workerUserID -> contractorID
}

func newMockDocRepo() *mockDocumentRepository {
	dc3ID := uuid.New()
	repseID := uuid.New()
	ineID := uuid.New()

	return &mockDocumentRepository{
		docs: make(map[uuid.UUID]*entity.Document),
		types: map[string]*entity.DocumentType{
			"DC3": {
				ID:        dc3ID,
				Code:      "DC3",
				Name:      "Constancia DC-3",
				AppliesTo: "worker",
			},
			"REPSE": {
				ID:        repseID,
				Code:      "REPSE",
				Name:      "Registro REPSE",
				AppliesTo: "contractor",
			},
			"INE": {
				ID:        ineID,
				Code:      "INE",
				Name:      "INE",
				AppliesTo: "worker",
			},
		},
		contractors: make(map[uuid.UUID]*entity.Contractor),
		workerLinks: make(map[uuid.UUID]uuid.UUID),
	}
}

func (m *mockDocumentRepository) Create(ctx context.Context, doc *entity.Document) error {
	m.docs[doc.ID] = doc
	return nil
}

func (m *mockDocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Document, error) {
	d, ok := m.docs[id]
	if !ok {
		return nil, domainErrors.ErrDocumentNotFound
	}
	return d, nil
}

func (m *mockDocumentRepository) ListByWorker(ctx context.Context, workerUserID uuid.UUID) ([]entity.Document, error) {
	var res []entity.Document
	for _, d := range m.docs {
		if d.UserID != nil && *d.UserID == workerUserID {
			res = append(res, *d)
		}
	}
	return res, nil
}

func (m *mockDocumentRepository) ListByContractor(ctx context.Context, contractorID uuid.UUID) ([]entity.Document, error) {
	var res []entity.Document
	for _, d := range m.docs {
		if d.ContractorID != nil && *d.ContractorID == contractorID {
			res = append(res, *d)
		}
	}
	return res, nil
}

func (m *mockDocumentRepository) GetDocumentTypeByCode(ctx context.Context, code string) (*entity.DocumentType, error) {
	t, ok := m.types[code]
	if !ok {
		return nil, domainErrors.ErrInvalidInput.WithMessage("Tipo de documento no encontrado.")
	}
	return t, nil
}

func (m *mockDocumentRepository) ListRequiredDocumentTypes(ctx context.Context) ([]entity.DocumentType, error) {
	var res []entity.DocumentType
	for _, t := range m.types {
		res = append(res, *t)
	}
	return res, nil
}

func (m *mockDocumentRepository) GetContractorByAdminUserID(ctx context.Context, adminUserID uuid.UUID) (*entity.Contractor, error) {
	for _, c := range m.contractors {
		if c.AdminUserID != nil && *c.AdminUserID == adminUserID {
			return c, nil
		}
	}
	return nil, domainErrors.ErrContractorNotFound
}

func (m *mockDocumentRepository) GetContractorByWorkerUserID(ctx context.Context, workerUserID uuid.UUID) (*entity.Contractor, error) {
	cID, ok := m.workerLinks[workerUserID]
	if !ok {
		return nil, domainErrors.ErrContractorNotFound
	}
	c, ok := m.contractors[cID]
	if !ok {
		return nil, domainErrors.ErrContractorNotFound
	}
	return c, nil
}

func (m *mockDocumentRepository) ListPendingReview(ctx context.Context, tenantID *uuid.UUID, statusFilter *entity.DocumentStatus) ([]entity.Document, error) {
	var res []entity.Document
	for _, d := range m.docs {
		if statusFilter != nil && d.Status != *statusFilter {
			continue
		}
		res = append(res, *d)
	}
	return res, nil
}

func (m *mockDocumentRepository) UpdateReviewStatus(ctx context.Context, id uuid.UUID, status entity.DocumentStatus, reviewedBy uuid.UUID, rejectionReason *string) (*entity.Document, error) {
	d, ok := m.docs[id]
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

func TestUploadDocumentUseCase_WorkerSuccess(t *testing.T) {
	repo := newMockDocRepo()
	store := newMockStorage()
	uc := document.NewUploadDocumentUseCase(repo, store)

	workerID := uuid.New()
	tenantID := uuid.New()
	ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(context.Background(), tenantID), workerID), "worker")

	pdfData := []byte("%PDF-1.4 valid test pdf file content")
	issuedAt := time.Now().AddDate(0, -1, 0)

	doc, err := uc.Execute(ctx, document.UploadDocumentInput{
		DocumentTypeCode: "DC3",
		FileName:         "dc3_certificate.pdf",
		File:             bytes.NewReader(pdfData),
		FileSize:         int64(len(pdfData)),
		IssuedAt:         &issuedAt,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc == nil {
		t.Fatal("expected created document, got nil")
	}
	if doc.Status != entity.DocumentStatusPendingReview {
		t.Errorf("expected status %s, got %s", entity.DocumentStatusPendingReview, doc.Status)
	}
	if doc.UserID == nil || *doc.UserID != workerID {
		t.Errorf("expected UserID %v, got %v", workerID, doc.UserID)
	}
	if doc.ContractorID != nil {
		t.Errorf("expected ContractorID nil, got %v", doc.ContractorID)
	}
	if _, ok := store.uploads[doc.R2Key]; !ok {
		t.Errorf("expected file to be uploaded at key %q", doc.R2Key)
	}
}

func TestUploadDocumentUseCase_WorkerForbiddenContractorDoc(t *testing.T) {
	repo := newMockDocRepo()
	store := newMockStorage()
	uc := document.NewUploadDocumentUseCase(repo, store)

	workerID := uuid.New()
	ctx := tenant.WithRole(tenant.WithUserID(context.Background(), workerID), "worker")

	pdfData := []byte("%PDF-1.4 test")
	_, err := uc.Execute(ctx, document.UploadDocumentInput{
		DocumentTypeCode: "REPSE", // contractor only doc
		FileName:         "repse.pdf",
		File:             bytes.NewReader(pdfData),
		FileSize:         int64(len(pdfData)),
	})

	if err == nil {
		t.Fatal("expected forbidden error for worker uploading REPSE, got nil")
	}
	if !errors.Is(err, domainErrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestUploadDocumentUseCase_InvalidMagicBytes(t *testing.T) {
	repo := newMockDocRepo()
	store := newMockStorage()
	uc := document.NewUploadDocumentUseCase(repo, store)

	workerID := uuid.New()
	ctx := tenant.WithRole(tenant.WithUserID(context.Background(), workerID), "worker")

	invalidData := []byte("This is plain text pretending to be PDF")
	_, err := uc.Execute(ctx, document.UploadDocumentInput{
		DocumentTypeCode: "DC3",
		FileName:         "fake.pdf",
		File:             bytes.NewReader(invalidData),
		FileSize:         int64(len(invalidData)),
	})

	if err == nil {
		t.Fatal("expected ErrInvalidFileType, got nil")
	}
	if !errors.Is(err, domainErrors.ErrInvalidFileType) {
		t.Errorf("expected ErrInvalidFileType, got %v", err)
	}
}

func TestUploadDocumentUseCase_FileTooLarge(t *testing.T) {
	repo := newMockDocRepo()
	store := newMockStorage()
	uc := document.NewUploadDocumentUseCase(repo, store)

	workerID := uuid.New()
	ctx := tenant.WithRole(tenant.WithUserID(context.Background(), workerID), "worker")

	pdfData := []byte("%PDF-1.4")
	_, err := uc.Execute(ctx, document.UploadDocumentInput{
		DocumentTypeCode: "DC3",
		FileName:         "huge.pdf",
		File:             bytes.NewReader(pdfData),
		FileSize:         11 * 1024 * 1024, // > 10MB
	})

	if err == nil {
		t.Fatal("expected ErrFileTooLarge, got nil")
	}
	if !errors.Is(err, domainErrors.ErrFileTooLarge) {
		t.Errorf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestDownloadDocumentUseCase_AuthorizedAndUnauthorized(t *testing.T) {
	repo := newMockDocRepo()
	store := newMockStorage()
	downloadUC := document.NewDownloadDocumentUseCase(repo, store)

	worker1ID := uuid.New()
	worker2ID := uuid.New()
	docID := uuid.New()
	key := "workers/w1/doc.pdf"
	content := []byte("%PDF-1.4 worker 1 document")
	store.uploads[key] = content

	repo.docs[docID] = &entity.Document{
		ID:       docID,
		UserID:   &worker1ID,
		R2Key:    key,
		FileName: "doc.pdf",
	}

	// Worker 1 can download
	ctx1 := tenant.WithRole(tenant.WithUserID(context.Background(), worker1ID), "worker")
	rc, doc, err := downloadUC.Execute(ctx1, docID)
	if err != nil {
		t.Fatalf("expected worker 1 to download, got error: %v", err)
	}
	defer rc.Close()
	if doc.ID != docID {
		t.Errorf("expected doc ID %v, got %v", docID, doc.ID)
	}

	// Worker 2 is forbidden
	ctx2 := tenant.WithRole(tenant.WithUserID(context.Background(), worker2ID), "worker")
	_, _, err = downloadUC.Execute(ctx2, docID)
	if err == nil {
		t.Fatal("expected worker 2 to get forbidden error, got nil")
	}
	if !errors.Is(err, domainErrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}

	// Guard is authorized
	guardID := uuid.New()
	ctxGuard := tenant.WithRole(tenant.WithUserID(context.Background(), guardID), "guard")
	rcGuard, _, err := downloadUC.Execute(ctxGuard, docID)
	if err != nil {
		t.Fatalf("expected guard to be authorized, got %v", err)
	}
	_ = rcGuard.Close()
}

func TestReviewDocumentUseCase_ApproveSuccess(t *testing.T) {
	repo := newMockDocRepo()
	uc := document.NewReviewDocumentUseCase(repo)

	ehsID := uuid.New()
	ctxEHS := tenant.WithRole(tenant.WithUserID(context.Background(), ehsID), "ehs_manager")

	docID := uuid.New()
	repo.docs[docID] = &entity.Document{
		ID:     docID,
		Status: entity.DocumentStatusPendingReview,
	}

	updated, err := uc.Execute(ctxEHS, document.ReviewDocumentInput{
		DocumentID: docID,
		Action:     "approve",
	})
	if err != nil {
		t.Fatalf("unexpected error approving document: %v", err)
	}
	if updated.Status != entity.DocumentStatusApproved {
		t.Errorf("expected status %s, got %s", entity.DocumentStatusApproved, updated.Status)
	}
	if updated.ReviewedBy == nil || *updated.ReviewedBy != ehsID {
		t.Errorf("expected reviewer %v, got %v", ehsID, updated.ReviewedBy)
	}
	if updated.RejectionReason != nil {
		t.Errorf("expected nil rejection reason, got %v", *updated.RejectionReason)
	}
}

func TestReviewDocumentUseCase_RejectSuccess(t *testing.T) {
	repo := newMockDocRepo()
	uc := document.NewReviewDocumentUseCase(repo)

	adminID := uuid.New()
	ctxAdmin := tenant.WithRole(tenant.WithUserID(context.Background(), adminID), "plant_admin")

	docID := uuid.New()
	repo.docs[docID] = &entity.Document{
		ID:     docID,
		Status: entity.DocumentStatusPendingReview,
	}

	reason := "El certificado DC-3 carece de la firma del capacitador registrado."
	updated, err := uc.Execute(ctxAdmin, document.ReviewDocumentInput{
		DocumentID:      docID,
		Action:          "reject",
		RejectionReason: reason,
	})
	if err != nil {
		t.Fatalf("unexpected error rejecting document: %v", err)
	}
	if updated.Status != entity.DocumentStatusRejected {
		t.Errorf("expected status %s, got %s", entity.DocumentStatusRejected, updated.Status)
	}
	if updated.RejectionReason == nil || *updated.RejectionReason != reason {
		t.Errorf("expected rejection reason %q, got %v", reason, updated.RejectionReason)
	}
}

func TestReviewDocumentUseCase_RejectMissingReasonFails(t *testing.T) {
	repo := newMockDocRepo()
	uc := document.NewReviewDocumentUseCase(repo)

	ehsID := uuid.New()
	ctxEHS := tenant.WithRole(tenant.WithUserID(context.Background(), ehsID), "ehs_manager")

	docID := uuid.New()
	repo.docs[docID] = &entity.Document{
		ID:     docID,
		Status: entity.DocumentStatusPendingReview,
	}

	_, err := uc.Execute(ctxEHS, document.ReviewDocumentInput{
		DocumentID:      docID,
		Action:          "reject",
		RejectionReason: "   ", // whitespace only
	})
	if err == nil {
		t.Fatal("expected error rejecting without reason, got nil")
	}
	if !errors.Is(err, domainErrors.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestReviewDocumentUseCase_WorkerRoleForbidden(t *testing.T) {
	repo := newMockDocRepo()
	uc := document.NewReviewDocumentUseCase(repo)

	workerID := uuid.New()
	ctxWorker := tenant.WithRole(tenant.WithUserID(context.Background(), workerID), "worker")

	docID := uuid.New()
	repo.docs[docID] = &entity.Document{
		ID:     docID,
		Status: entity.DocumentStatusPendingReview,
	}

	_, err := uc.Execute(ctxWorker, document.ReviewDocumentInput{
		DocumentID: docID,
		Action:     "approve",
	})
	if err == nil {
		t.Fatal("expected forbidden error for worker, got nil")
	}
	if !errors.Is(err, domainErrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestListPendingDocumentsUseCase_Success(t *testing.T) {
	repo := newMockDocRepo()
	uc := document.NewListPendingDocumentsUseCase(repo)

	ehsID := uuid.New()
	ctxEHS := tenant.WithRole(tenant.WithUserID(context.Background(), ehsID), "ehs_manager")

	doc1 := uuid.New()
	doc2 := uuid.New()
	repo.docs[doc1] = &entity.Document{ID: doc1, Status: entity.DocumentStatusPendingReview}
	repo.docs[doc2] = &entity.Document{ID: doc2, Status: entity.DocumentStatusApproved}

	// 1. List pending only
	docs, err := uc.Execute(ctxEHS, "pending_review")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 1 || docs[0].ID != doc1 {
		t.Errorf("expected only pending doc, got %d docs", len(docs))
	}

	// 2. List all
	allDocs, err := uc.Execute(ctxEHS, "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(allDocs) != 2 {
		t.Errorf("expected 2 docs, got %d", len(allDocs))
	}
}
