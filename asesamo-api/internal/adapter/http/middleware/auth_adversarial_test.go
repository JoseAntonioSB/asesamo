package middleware_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"asesamo-api/internal/adapter/http/dto"
	"asesamo-api/internal/adapter/http/middleware"
	"asesamo-api/pkg/jwt"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdversarial_AuthMiddleware(t *testing.T) {
	const secret = "a_very_secure_test_secret_32_bytes_long_ok!!"
	userID := uuid.New()
	tenantID := uuid.New()

	handler := middleware.Auth(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("passed"))
	}))

	t.Run("ATTACK: alg 'none' token via HTTP Authorization header returns 401 INVALID_TOKEN", func(t *testing.T) {
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
		claimsJSON := fmt.Sprintf(`{"sub":"%s","tenant_id":"%s","role":"super_admin","exp":%d}`,
			userID.String(), tenantID.String(), time.Now().Add(time.Hour).Unix())
		payload := base64.RawURLEncoding.EncodeToString([]byte(claimsJSON))
		noneToken := fmt.Sprintf("%s.%s.", header, payload)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+noneToken)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var resp dto.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "INVALID_TOKEN", resp.Error.Code)
	})

	t.Run("ATTACK: RS256 token via HTTP Authorization header returns 401 INVALID_TOKEN", func(t *testing.T) {
		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		rsaClaims := jwt.Claims{
			TenantID: tenantID,
			Role:     "super_admin",
			RegisteredClaims: jwtlib.RegisteredClaims{
				Subject:   userID.String(),
				ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
		rsaTokenObj := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, rsaClaims)
		rsaTokenStr, err := rsaTokenObj.SignedString(rsaKey)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+rsaTokenStr)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var resp dto.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "INVALID_TOKEN", resp.Error.Code)
	})

	t.Run("ATTACK: Whitespace injection and multiple spaces in Authorization header", func(t *testing.T) {
		token, err := jwt.SignToken(userID, tenantID, "guard", []byte(secret))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		// Header with extra spaces between Bearer and token
		req.Header.Set("Authorization", "Bearer   "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "strings.TrimSpace(parts[1]) should handle leading space cleanly")
	})

	t.Run("ATTACK: Case sensitivity on bearer prefix", func(t *testing.T) {
		token, err := jwt.SignToken(userID, tenantID, "guard", []byte(secret))
		require.NoError(t, err)

		// lowercase bearer
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "strings.EqualFold accepts 'bearer'")
	})
}
