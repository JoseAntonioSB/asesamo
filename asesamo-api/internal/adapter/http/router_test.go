package http_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	adapterHTTP "asesamo-api/internal/adapter/http"
	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/usecase/access"
	"asesamo-api/internal/core/usecase/auth"
	"asesamo-api/internal/core/usecase/document"
	"asesamo-api/pkg/jwt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type RouterMockUserRepo struct{ mock.Mock }

func (m *RouterMockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	return nil, args.Error(1)
}
func (m *RouterMockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *RouterMockUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	return nil
}
func (m *RouterMockUserRepo) UpdateThemePreference(ctx context.Context, id uuid.UUID, theme string) error {
	return nil
}


type RouterMockAccessPointRepo struct{ mock.Mock }

func (m *RouterMockAccessPointRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
	return nil, nil
}
func (m *RouterMockAccessPointRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}
func (m *RouterMockAccessPointRepo) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
	return nil, nil
}

type RouterMockDocRepo struct{}

func (r *RouterMockDocRepo) Create(ctx context.Context, doc *entity.Document) error {
	return nil
}
func (r *RouterMockDocRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Document, error) {
	return nil, nil
}
func (r *RouterMockDocRepo) ListByWorker(ctx context.Context, workerUserID uuid.UUID) ([]entity.Document, error) {
	return []entity.Document{}, nil
}
func (r *RouterMockDocRepo) ListByContractor(ctx context.Context, contractorID uuid.UUID) ([]entity.Document, error) {
	return []entity.Document{}, nil
}
func (r *RouterMockDocRepo) GetDocumentTypeByCode(ctx context.Context, code string) (*entity.DocumentType, error) {
	return nil, nil
}
func (r *RouterMockDocRepo) ListRequiredDocumentTypes(ctx context.Context) ([]entity.DocumentType, error) {
	return []entity.DocumentType{}, nil
}
func (r *RouterMockDocRepo) GetContractorByAdminUserID(ctx context.Context, adminUserID uuid.UUID) (*entity.Contractor, error) {
	return nil, nil
}
func (r *RouterMockDocRepo) GetContractorByWorkerUserID(ctx context.Context, workerUserID uuid.UUID) (*entity.Contractor, error) {
	return nil, nil
}
func (r *RouterMockDocRepo) ListPendingReview(ctx context.Context, tenantID *uuid.UUID, statusFilter *entity.DocumentStatus) ([]entity.Document, error) {
	return []entity.Document{}, nil
}
func (r *RouterMockDocRepo) UpdateReviewStatus(ctx context.Context, id uuid.UUID, status entity.DocumentStatus, reviewedBy uuid.UUID, rejectionReason *string) (*entity.Document, error) {
	return nil, nil
}

type RouterMockStorage struct{}

func (s *RouterMockStorage) Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error {
	return nil
}
func (s *RouterMockStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, nil
}
func (s *RouterMockStorage) Delete(ctx context.Context, key string) error {
	return nil
}
func (s *RouterMockStorage) GetURL(ctx context.Context, key string) (string, error) {
	return "", nil
}

const routerTestSecret = "super_secret_jwt_key_min_32_characters_long_12345"

func TestRouter(t *testing.T) {
	userRepo := new(RouterMockUserRepo)
	apRepo := new(RouterMockAccessPointRepo)
	docRepo := &RouterMockDocRepo{}
	storageSvc := &RouterMockStorage{}

	healthHandler := handler.NewHealthHandler(nil, "1.0.0")
	authHandler := handler.NewAuthHandler(auth.NewLoginUseCase(userRepo, routerTestSecret))
	apHandler := handler.NewAccessPointHandler(access.NewListAccessPointsUseCase(apRepo))

	uploadUC := document.NewUploadDocumentUseCase(docRepo, storageSvc)
	listMyDocsUC := document.NewListMyDocumentsUseCase(docRepo)
	downloadUC := document.NewDownloadDocumentUseCase(docRepo, storageSvc)
	listDocTypesUC := document.NewListDocumentTypesUseCase(docRepo)
	docHandler := handler.NewDocumentHandler(uploadUC, listMyDocsUC, downloadUC, listDocTypesUC)

	r := adapterHTTP.NewRouter(adapterHTTP.RouterConfig{
		HealthHandler:      healthHandler,
		AuthHandler:        authHandler,
		AccessPointHandler: apHandler,
		DocumentHandler:    docHandler,
		JWTSecret:          routerTestSecret,
		CORSOrigins:        []string{"*"},
	})

	t.Run("public GET /health succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("public GET /healthz succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("public GET /api/v1/health succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("unauthenticated GET /api/v1/access-points returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("authenticated GET /api/v1/access-points succeeds", func(t *testing.T) {
		tenantID := uuid.New()
		userID := uuid.New()
		token, err := jwt.SignToken(userID, tenantID, "guard", []byte(routerTestSecret))
		require.NoError(t, err)

		apRepo.On("ListByTenant", mock.Anything, tenantID).Return([]entity.AccessPoint{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("unauthenticated POST /api/v1/documents/upload returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/upload", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("unauthenticated GET /api/v1/documents/my-documents returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/my-documents", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("unauthenticated GET /api/v1/documents/uuid/download returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/"+uuid.New().String()+"/download", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("authenticated GET /api/v1/documents/my-documents succeeds", func(t *testing.T) {
		tenantID := uuid.New()
		userID := uuid.New()
		token, err := jwt.SignToken(userID, tenantID, "worker", []byte(routerTestSecret))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/my-documents", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
