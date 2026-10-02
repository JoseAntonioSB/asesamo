package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"asesamo-api/internal/adapter/http/dto"
	"asesamo-api/internal/adapter/http/middleware"
	"asesamo-api/pkg/jwt"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "super_secret_jwt_key_min_32_characters_long_12345"

func TestAuthMiddleware(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	role := "guard"

	authHandler := middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxTenant, err := tenant.FromContext(r.Context())
		assert.NoError(t, err)
		assert.Equal(t, tenantID, ctxTenant)

		ctxUser, err := tenant.UserIDFromContext(r.Context())
		assert.NoError(t, err)
		assert.Equal(t, userID, ctxUser)

		ctxRole, err := tenant.RoleFromContext(r.Context())
		assert.NoError(t, err)
		assert.Equal(t, role, ctxRole)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	t.Run("valid Bearer token allows request and propagates context", func(t *testing.T) {
		token, err := jwt.SignToken(userID, tenantID, role, []byte(testSecret))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		authHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})

	t.Run("missing Authorization header returns 401 UNAUTHORIZED", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()

		authHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var resp dto.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
	})

	t.Run("malformed header (missing Bearer prefix) returns 401 UNAUTHORIZED", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Token some-token-string")
		rec := httptest.NewRecorder()

		authHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var resp dto.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
	})

	t.Run("expired token returns 401 TOKEN_EXPIRED", func(t *testing.T) {
		token, err := jwt.SignTokenWithDuration(userID, tenantID, role, []byte(testSecret), -5*time.Minute)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		authHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var resp dto.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "TOKEN_EXPIRED", resp.Error.Code)
	})

	t.Run("tampered token returns 401 INVALID_TOKEN", func(t *testing.T) {
		token, err := jwt.SignToken(userID, tenantID, role, []byte(testSecret))
		require.NoError(t, err)

		tampered := token[:len(token)-6] + "xxxxxx"
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tampered)
		rec := httptest.NewRecorder()

		authHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var resp dto.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "INVALID_TOKEN", resp.Error.Code)
	})
}

func TestRequireRole(t *testing.T) {
	guardOnly := middleware.RequireRole("guard")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("access granted"))
	}))

	t.Run("matching role succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/guard-only", nil)
		ctx := tenant.WithRole(req.Context(), "guard")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		guardOnly.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "access granted", rec.Body.String())
	})

	t.Run("non-matching role returns 403 FORBIDDEN", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/guard-only", nil)
		ctx := tenant.WithRole(req.Context(), "worker")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		guardOnly.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
		var resp dto.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "FORBIDDEN", resp.Error.Code)
	})

	t.Run("missing role returns 401 UNAUTHORIZED", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/guard-only", nil)
		rec := httptest.NewRecorder()

		guardOnly.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
