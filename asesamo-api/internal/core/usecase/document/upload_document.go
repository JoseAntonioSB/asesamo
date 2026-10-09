package document

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"asesamo-api/internal/adapter/storage"
	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

// UploadDocumentInput defines the parameters for uploading a document.
type UploadDocumentInput struct {
	DocumentTypeCode string
	FileName         string
	File             io.Reader
	FileSize         int64
	IssuedAt         *time.Time
	ExpiresAt        *time.Time
	WorkerID         *uuid.UUID
	ContractorID     *uuid.UUID
}

// UploadDocumentUseCase handles document uploading, validation, and storage.
type UploadDocumentUseCase struct {
	docRepo        repository.DocumentRepository
	storageService repository.StorageService
}

// NewUploadDocumentUseCase constructs a new UploadDocumentUseCase.
func NewUploadDocumentUseCase(
	docRepo repository.DocumentRepository,
	storageService repository.StorageService,
) *UploadDocumentUseCase {
	return &UploadDocumentUseCase{
		docRepo:        docRepo,
		storageService: storageService,
	}
}

// Execute validates and uploads a document.
func (uc *UploadDocumentUseCase) Execute(ctx context.Context, input UploadDocumentInput) (*entity.Document, error) {
	// 1. Authenticated user and role
	authUserID, err := tenant.UserIDFromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrUnauthorized
	}
	role, err := tenant.RoleFromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrUnauthorized
	}

	// 2. Validate input and file size
	if input.FileSize <= 0 {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El archivo no contiene datos.")
	}
	if input.FileSize > storage.MaxFileSize {
		return nil, domainErrors.ErrFileTooLarge
	}
	if strings.TrimSpace(input.DocumentTypeCode) == "" {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El código de tipo de documento es obligatorio.")
	}

	// 3. Inspect magic bytes for authentic PDF, JPEG, or PNG
	headerBuf := make([]byte, 512)
	n, err := io.ReadFull(input.File, headerBuf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, domainErrors.ErrInvalidInput.WithMessage("No se pudo leer el archivo cargado.")
	}
	headerBuf = headerBuf[:n]

	detectedMime, err := storage.ValidateMagicBytes(headerBuf)
	if err != nil {
		return nil, err
	}

	// Combine header with the remaining stream
	fullReader := io.MultiReader(bytes.NewReader(headerBuf), input.File)

	// 4. Validate document type from catalog
	docType, err := uc.docRepo.GetDocumentTypeByCode(ctx, strings.ToUpper(strings.TrimSpace(input.DocumentTypeCode)))
	if err != nil {
		return nil, err
	}

	// 5. Determine document ownership based on role and applies_to
	var targetUserID *uuid.UUID
	var targetContractorID *uuid.UUID

	switch role {
	case "worker":
		if docType.AppliesTo != "worker" {
			return nil, domainErrors.ErrForbidden.WithMessage("Los trabajadores solo pueden cargar documentos personales (applies_to: worker).")
		}
		targetUserID = &authUserID

	case "contractor_admin":
		if docType.AppliesTo == "contractor" {
			contractor, err := uc.docRepo.GetContractorByAdminUserID(ctx, authUserID)
			if err != nil {
				return nil, err
			}
			targetContractorID = &contractor.ID
		} else {
			// Worker document uploaded by contractor admin
			if input.WorkerID == nil {
				return nil, domainErrors.ErrInvalidInput.WithMessage("Debe especificar el worker_id del trabajador para este tipo de documento.")
			}
			contractor, err := uc.docRepo.GetContractorByWorkerUserID(ctx, *input.WorkerID)
			if err != nil {
				return nil, err
			}
			adminContractor, err := uc.docRepo.GetContractorByAdminUserID(ctx, authUserID)
			if err != nil {
				return nil, err
			}
			if contractor.ID != adminContractor.ID {
				return nil, domainErrors.ErrForbidden.WithMessage("El trabajador no pertenece a su empresa contratista.")
			}
			targetUserID = input.WorkerID
		}

	case "plant_admin", "super_admin", "ehs_manager":
		if docType.AppliesTo == "worker" {
			if input.WorkerID != nil {
				targetUserID = input.WorkerID
			} else {
				targetUserID = &authUserID
			}
		} else {
			if input.ContractorID == nil {
				return nil, domainErrors.ErrInvalidInput.WithMessage("Debe especificar el contractor_id para documentos de empresa contratista.")
			}
			targetContractorID = input.ContractorID
		}

	default:
		return nil, domainErrors.ErrForbidden.WithMessage("Su rol no tiene autorización para cargar documentos.")
	}

	// Check DB ownership constraint: exactly one must be set
	if (targetUserID == nil && targetContractorID == nil) || (targetUserID != nil && targetContractorID != nil) {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El documento debe pertenecer exclusivamente a un trabajador o a un contratista.")
	}

	// 6. Build file key according to standard: {tenant_or_global}/{entity_type}/{id}/{uuid}_{filename}
	tenantID, _ := tenant.FromContext(ctx)
	tenantPrefix := "global"
	if tenantID != uuid.Nil {
		tenantPrefix = tenantID.String()
	}

	entityType := "workers"
	entityID := ""
	if targetUserID != nil {
		entityID = targetUserID.String()
	} else {
		entityType = "contractors"
		entityID = targetContractorID.String()
	}

	fileUUID := uuid.New().String()
	ext := strings.ToLower(filepath.Ext(input.FileName))
	if ext == "" {
		switch detectedMime {
		case storage.ContentTypePDF:
			ext = ".pdf"
		case storage.ContentTypeJPEG:
			ext = ".jpg"
		case storage.ContentTypePNG:
			ext = ".png"
		}
	}
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	r2Key := fmt.Sprintf("%s/%s/%s/%s%s", tenantPrefix, entityType, entityID, fileUUID, ext)

	// 7. Persist to storage adapter
	if err := uc.storageService.Upload(ctx, r2Key, fullReader, input.FileSize, detectedMime); err != nil {
		return nil, err
	}

	// 8. Auto-calculate expiration date if omitted and default validity is set
	expiresAt := input.ExpiresAt
	if expiresAt == nil && input.IssuedAt != nil && docType.DefaultValidityDays != nil {
		calcExp := input.IssuedAt.AddDate(0, 0, *docType.DefaultValidityDays)
		expiresAt = &calcExp
	}

	fileSizeInt := int(input.FileSize)
	doc := &entity.Document{
		ID:               uuid.New(),
		UserID:           targetUserID,
		ContractorID:     targetContractorID,
		DocumentTypeID:   docType.ID,
		DocumentTypeCode: docType.Code,
		DocumentTypeName: docType.Name,
		R2Key:            r2Key,
		FileName:         filepath.Base(input.FileName),
		FileSizeBytes:    &fileSizeInt,
		MimeType:         &detectedMime,
		IssuedAt:         input.IssuedAt,
		ExpiresAt:        expiresAt,
		Status:           entity.DocumentStatusPendingReview,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	// 9. Persist document metadata in DB
	if err := uc.docRepo.Create(ctx, doc); err != nil {
		// Attempt rollback in storage if DB fails
		_ = uc.storageService.Delete(ctx, r2Key)
		return nil, err
	}

	return doc, nil
}
