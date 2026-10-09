# Project: ASESAMO Document Upload Module

## Architecture
Clean Architecture with strict boundary separation between domains, ports, and adapters.

- **Backend (`asesamo-api`)**:
  - `internal/core/domain/storage/`: Storage abstraction port (`StorageService` interface: `Upload`, `Download`, `GetURL`, `Delete`).
  - `internal/adapter/storage/`: Concrete storage adapters (`LocalStorageService` for local filesystem / docker volume, Cloudflare R2 / S3 ready). Enforces magic bytes validation (PDF `%PDF-`, JPEG `\xFF\xD8\xFF`, PNG `\x89PNG`), 10 MB file cap, and path sanitization.
  - `internal/core/domain/entity/`: Pure domain entities (`Document`, `DocumentType`, `User`, `Contractor`).
  - `internal/core/domain/repository/`: Domain repository ports (`DocumentRepository`).
  - `internal/adapter/postgres/`: PostgreSQL repository implementations using parameterized SQL (`database/sql` + `pgx/v5`).
  - `internal/core/usecase/document/`: Domain use cases (`UploadDocumentUseCase`, `ListMyDocumentsUseCase`, `DownloadDocumentUseCase`).
  - `internal/adapter/http/handler/`: HTTP handlers with Spanish error envelopes (`DocumentHandler`).
  - `internal/adapter/http/router.go`: Chi router wiring with JWT auth and role/tenant context.

- **Frontend (`asesamo-web`)**:
  - Next.js 14 App Router (`app/documents/page.tsx`).
  - Edge route protection via `middleware.ts` for `/documents`.
  - API Client (`lib/api/client.ts`): Multipart form-data support with automatic boundary detection and typed document endpoints.
  - UI Components: Accessible upload dialog, drag-and-drop dropzone with client validation, document status badges, Mexican regulatory catalog selector.
  - Sonner toast notifications for user feedback.

## Feature Inventory
| # | Feature | Description | Milestone | Source |
|---|---------|-------------|-----------|--------|
| 1 | Storage Port Abstraction | Domain interface `StorageService` (`Upload`, `Download`, `GetURL`, `Delete`) | M1 | ORIGINAL_REQUEST §R1 |
| 2 | Local Storage Adapter | `LocalStorageService` with directory isolation and path traversal prevention | M1 | ORIGINAL_REQUEST §R1 |
| 3 | Magic Bytes & Size Validation | Header sniffing for PDF, JPEG, PNG binary signatures and 10MB size limit | M1 | ORIGINAL_REQUEST §R1 |
| 4 | Storage Key Normalization | Standardized `{tenant_id_or_global}/{entity_type}/{id}/{filename_uuid}.{ext}` | M1 | ORIGINAL_REQUEST §R1 |
| 5 | Document Repository Port Extension | Adding `Create`, `GetDocumentTypeByCode`, `ListDocumentTypes`, etc. | M2 | ORIGINAL_REQUEST §R2 |
| 6 | Postgres Document Repository | PostgreSQL adapter for document persistence and queries | M2 | ORIGINAL_REQUEST §R2 |
| 7 | Document Use Cases | `UploadDocumentUseCase`, `ListMyDocumentsUseCase`, `DownloadDocumentUseCase` | M2 | ORIGINAL_REQUEST §R2 |
| 8 | REST Upload Endpoint | `POST /api/v1/documents/upload` with multipart parsing and Spanish errors | M2 | ORIGINAL_REQUEST §R2 |
| 9 | REST My-Documents Endpoint | `GET /api/v1/documents/my-documents` returning `{ "data": ... }` | M2 | ORIGINAL_REQUEST §R2 |
| 10 | REST Download Endpoint | `GET /api/v1/documents/{id}/download` with streaming and ownership checks | M2 | ORIGINAL_REQUEST §R2 |
| 11 | Router & Dependency Wiring | Integration into `router.go`, `config.go`, and `cmd/api/main.go` | M2 | ORIGINAL_REQUEST §R2 |
| 12 | Frontend API Multipart Fix | Removing hardcoded JSON Content-Type for `FormData` in `ApiClient` | M3 | Frontend Survey |
| 13 | Frontend API Document Methods | Adding `uploadDocument`, `getMyDocuments`, and `getDocumentDownloadUrl` | M3 | ORIGINAL_REQUEST §R3 |
| 14 | Toast Notifications (Sonner) | Installing `sonner`, mounting `<Toaster />`, showing feedback | M3 | ORIGINAL_REQUEST §R3 |
| 15 | Route & Middleware Protection | Establishing `/documents` and protecting in `middleware.ts` | M3 | Frontend Survey |
| 16 | Document Management View | Responsive document list/grid with status badges (`pending_review`, etc.) | M3 | ORIGINAL_REQUEST §R3 |
| 17 | Accessible Upload Dialog | Modal with document type selector, date pickers, dropzone | M3 | ORIGINAL_REQUEST §R3 |
| 18 | Client-side File Validation | Validation for 10MB limit and PDF/JPEG/PNG before upload | M3 | ORIGINAL_REQUEST §R3 |
| 19 | E2E Test Suite (Tiers 1-4) | Comprehensive opaque-box test runner covering all acceptance criteria | M4 | Dual-Track Testing |
| 20 | Adversarial Coverage Hardening | Tier 5 white-box coverage testing & final forensic audit | M5 | Final Milestone |

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| M1 | Backend Storage Adapter & Clean Domain | StorageService interface, LocalStorageService, magic byte validator, 10MB cap | none | PLANNED |
| M2 | Backend Document Upload & Management Endpoints | Postgres repository, document use cases, HTTP handlers, routes, backend unit tests | M1 | PLANNED |
| M3 | Frontend Document Upload UI | ApiClient multipart fix, Sonner toast, `/documents` page, upload dialog, validation | M2 | PLANNED |
| M4 | E2E Test Verification (Tiers 1-4) | Verification of 100% passing E2E tests against live backend and frontend | M2, M3, E2E-Track | PLANNED |
| M5 | Adversarial Hardening (Tier 5) & Victory Audit | Adversarial test suite, integrity audit verification, sentinel handoff | M4 | PLANNED |

## Interface Contracts
### Domain Storage Port (`internal/core/domain/storage/storage.go`)
```go
package storage

import (
    "context"
    "io"
)

type StorageService interface {
    Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) (string, error)
    Download(ctx context.Context, key string) (io.ReadCloser, string, error)
    GetURL(ctx context.Context, key string) (string, error)
    Delete(ctx context.Context, key string) error
}
```

### Document Upload REST Contract (`POST /api/v1/documents/upload`)
- **Request**: `multipart/form-data`
  - `file`: binary stream
  - `document_type_code`: string (`"DC3"`, `"REPSE"`, `"SUA"`, `"INE"`, `"MEDICAL"`, `"SAT_32D"`)
  - `issued_at`: string (`YYYY-MM-DD`)
  - `expires_at`: optional string (`YYYY-MM-DD`)
- **Response** (HTTP 201 Created):
  ```json
  {
    "data": {
      "id": "uuid",
      "user_id": "uuid|null",
      "contractor_id": "uuid|null",
      "document_type_id": "uuid",
      "document_type_code": "DC3",
      "file_name": "dc3.pdf",
      "file_size_bytes": 1048576,
      "mime_type": "application/pdf",
      "issued_at": "2026-01-15",
      "expires_at": "2027-01-15",
      "status": "pending_review",
      "created_at": "2026-10-08T21:40:00Z"
    }
  }
  ```
- **Error Response** (HTTP 400/422):
  ```json
  {
    "error": {
      "code": "INVALID_FILE_TYPE",
      "message": "Tipo de archivo no permitido. Solo se aceptan formatos PDF, JPEG y PNG."
    }
  }
  ```

### Document Listing REST Contract (`GET /api/v1/documents/my-documents`)
- **Headers**: `Authorization: Bearer <jwt>`
- **Response** (HTTP 200 OK):
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "document_type_code": "DC3",
        "document_type_name": "Constancia de Capacitación DC-3",
        "file_name": "dc3.pdf",
        "status": "pending_review",
        "issued_at": "2026-01-15",
        "expires_at": "2027-01-15",
        "rejection_reason": null,
        "created_at": "2026-10-08T21:40:00Z"
      }
    ]
  }
  ```

### Document Download REST Contract (`GET /api/v1/documents/{id}/download`)
- **Headers**: `Authorization: Bearer <jwt>`
- **Response**: Binary stream with headers:
  - `Content-Type: application/pdf` (or `image/jpeg`, `image/png`)
  - `Content-Disposition: inline; filename="original_filename.pdf"`

## Code Layout
### Backend (`asesamo-api/`)
- `internal/core/domain/storage/storage.go`: Storage interface
- `internal/adapter/storage/local_storage.go`: Local disk storage adapter
- `internal/adapter/storage/local_storage_test.go`: Storage adapter tests
- `internal/core/domain/errors/errors.go`: Spanish domain errors
- `internal/core/domain/repository/document_repository.go`: Extended repository interface
- `internal/adapter/postgres/document_repository.go`: Postgres document repository
- `internal/core/usecase/document/upload.go`: Upload use case
- `internal/core/usecase/document/list_my_documents.go`: My-documents usecase
- `internal/core/usecase/document/download.go`: Download usecase
- `internal/adapter/http/dto/document_dto.go`: DTO structures and JSON envelopes
- `internal/adapter/http/handler/document_handler.go`: HTTP handler implementation
- `internal/adapter/http/router.go`: Route registration
- `internal/infra/config/config.go`: Storage configuration
- `cmd/api/main.go`: Dependency injection and server boot

### Frontend (`asesamo-web/`)
- `middleware.ts`: Edge route guard for `/documents`
- `package.json`: Dependency addition (`sonner`)
- `app/layout.tsx`: Sonner `<Toaster />` integration
- `lib/api/types.ts`: TypeScript document and upload interfaces
- `lib/api/client.ts`: Multipart FormData support & document client methods
- `app/documents/page.tsx`: Document management portal page
- `components/documents/upload-dialog.tsx`: Upload modal
- `components/documents/file-dropzone.tsx`: Drag-and-drop zone with client validation
- `components/documents/document-list.tsx`: Required documents table
- `components/documents/document-status-badge.tsx`: Document status badges
- `components/documents/document-type-select.tsx`: Document type selector
