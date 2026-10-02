package jwt_test

import (
	"testing"
	"time"

	"asesamo-api/pkg/jwt"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testSecret = []byte("super_secret_jwt_key_min_32_characters_long_12345")

func TestSignAndValidateToken(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	role := "guard"

	t.Run("successfully signs and validates token", func(t *testing.T) {
		tokenStr, err := jwt.SignToken(userID, tenantID, role, testSecret)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenStr)

		claims, err := jwt.ValidateToken(tokenStr, testSecret)
		require.NoError(t, err)
		require.NotNil(t, claims)

		assert.Equal(t, userID.String(), claims.Subject)
		assert.Equal(t, tenantID, claims.TenantID)
		assert.Equal(t, role, claims.Role)

		parsedUserID, err := claims.UserID()
		require.NoError(t, err)
		assert.Equal(t, userID, parsedUserID)
	})

	t.Run("rejects token with secret too short", func(t *testing.T) {
		shortSecret := []byte("short_secret")
		_, err := jwt.SignToken(userID, tenantID, role, shortSecret)
		assert.ErrorIs(t, err, jwt.ErrSecretTooShort)

		_, err = jwt.ValidateToken("some.token.here", shortSecret)
		assert.ErrorIs(t, err, jwt.ErrSecretTooShort)
	})

	t.Run("rejects token with empty secret", func(t *testing.T) {
		_, err := jwt.SignToken(userID, tenantID, role, nil)
		assert.ErrorIs(t, err, jwt.ErrEmptySecret)

		_, err = jwt.ValidateToken("some.token.here", nil)
		assert.ErrorIs(t, err, jwt.ErrEmptySecret)
	})

	t.Run("rejects expired token", func(t *testing.T) {
		// Sign token with negative duration (already expired)
		tokenStr, err := jwt.SignTokenWithDuration(userID, tenantID, role, testSecret, -10*time.Minute)
		require.NoError(t, err)

		claims, err := jwt.ValidateToken(tokenStr, testSecret)
		assert.ErrorIs(t, err, jwt.ErrExpiredToken)
		assert.Nil(t, claims)
	})

	t.Run("rejects tampered token signature", func(t *testing.T) {
		tokenStr, err := jwt.SignToken(userID, tenantID, role, testSecret)
		require.NoError(t, err)

		// Tamper with last characters of token
		tamperedToken := tokenStr[:len(tokenStr)-5] + "XXXXX"
		claims, err := jwt.ValidateToken(tamperedToken, testSecret)
		assert.ErrorIs(t, err, jwt.ErrInvalidToken)
		assert.Nil(t, claims)
	})

	t.Run("rejects token signed with different secret", func(t *testing.T) {
		tokenStr, err := jwt.SignToken(userID, tenantID, role, testSecret)
		require.NoError(t, err)

		differentSecret := []byte("another_different_secret_key_with_32_plus_chars")
		claims, err := jwt.ValidateToken(tokenStr, differentSecret)
		assert.ErrorIs(t, err, jwt.ErrInvalidToken)
		assert.Nil(t, claims)
	})

	t.Run("rejects malformed token string", func(t *testing.T) {
		claims, err := jwt.ValidateToken("not-a-valid-jwt-token", testSecret)
		assert.ErrorIs(t, err, jwt.ErrInvalidToken)
		assert.Nil(t, claims)
	})

	t.Run("claims UserID returns error for empty or invalid UUID", func(t *testing.T) {
		claimsEmpty := jwt.Claims{
			RegisteredClaims: jwtlib.RegisteredClaims{Subject: ""},
		}
		_, err := claimsEmpty.UserID()
		assert.ErrorIs(t, err, jwt.ErrInvalidSubject)

		claimsInvalid := jwt.Claims{
			RegisteredClaims: jwtlib.RegisteredClaims{Subject: "not-a-uuid"},
		}
		_, err = claimsInvalid.UserID()
		assert.ErrorIs(t, err, jwt.ErrInvalidSubject)
	})
}
