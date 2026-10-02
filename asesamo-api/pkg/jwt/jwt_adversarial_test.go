package jwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	"asesamo-api/pkg/jwt"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdversarial_JWT_Security(t *testing.T) {
	secret := []byte("a_very_secure_test_secret_32_bytes_long_ok!!")
	userID := uuid.New()
	tenantID := uuid.New()

	t.Run("ATTACK: alg 'none' rejection (unsigned token with trailing dot)", func(t *testing.T) {
		// Header: {"alg":"none","typ":"JWT"} -> eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
		claimsJSON := fmt.Sprintf(`{"sub":"%s","tenant_id":"%s","role":"super_admin","exp":%d}`,
			userID.String(), tenantID.String(), time.Now().Add(time.Hour).Unix())
		payload := base64.RawURLEncoding.EncodeToString([]byte(claimsJSON))
		noneToken := fmt.Sprintf("%s.%s.", header, payload)

		claims, err := jwt.ValidateToken(noneToken, secret)
		assert.Error(t, err, "alg=none token must be rejected")
		assert.Nil(t, claims)
	})

	t.Run("ATTACK: alg 'none' rejection without trailing dot", func(t *testing.T) {
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
		claimsJSON := fmt.Sprintf(`{"sub":"%s","tenant_id":"%s","role":"super_admin","exp":%d}`,
			userID.String(), tenantID.String(), time.Now().Add(time.Hour).Unix())
		payload := base64.RawURLEncoding.EncodeToString([]byte(claimsJSON))
		noneToken := fmt.Sprintf("%s.%s", header, payload)

		claims, err := jwt.ValidateToken(noneToken, secret)
		assert.Error(t, err, "alg=none token without dot must be rejected")
		assert.Nil(t, claims)
	})

	t.Run("ATTACK: alg 'None' and 'NONE' case variants rejection", func(t *testing.T) {
		variants := []string{"None", "NONE", "nOnE"}
		for _, variant := range variants {
			header := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"alg":"%s","typ":"JWT"}`, variant)))
			claimsJSON := fmt.Sprintf(`{"sub":"%s","tenant_id":"%s","role":"super_admin","exp":%d}`,
				userID.String(), tenantID.String(), time.Now().Add(time.Hour).Unix())
			payload := base64.RawURLEncoding.EncodeToString([]byte(claimsJSON))
			token := fmt.Sprintf("%s.%s.", header, payload)

			claims, err := jwt.ValidateToken(token, secret)
			assert.Error(t, err, "cased none token %q must be rejected", variant)
			assert.Nil(t, claims)
		}
	})

	t.Run("ATTACK: alg mismatch - RS256 token presented to HS256 validator", func(t *testing.T) {
		// Generate an RSA private key to create a validly signed RS256 token
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

		// Attempt validation against HS256 secret
		claims, err := jwt.ValidateToken(rsaTokenStr, secret)
		assert.Error(t, err, "RS256 token must be rejected by HS256 validator")
		assert.ErrorIs(t, err, jwt.ErrInvalidToken)
		assert.Nil(t, claims)
	})

	t.Run("ATTACK: alg mismatch - HS384 and HS512 rejected when HS256 is expected", func(t *testing.T) {
		for _, method := range []*jwtlib.SigningMethodHMAC{jwtlib.SigningMethodHS384, jwtlib.SigningMethodHS512} {
			claims := jwt.Claims{
				TenantID: tenantID,
				Role:     "guard",
				RegisteredClaims: jwtlib.RegisteredClaims{
					Subject:   userID.String(),
					ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
				},
			}
			tkn := jwtlib.NewWithClaims(method, claims)
			signed, err := tkn.SignedString(secret)
			require.NoError(t, err)

			valClaims, err := jwt.ValidateToken(signed, secret)
			assert.Error(t, err, "Token with alg %s should be rejected", method.Alg())
			assert.Nil(t, valClaims)
		}
	})

	t.Run("ATTACK: Privilege escalation via payload tampering", func(t *testing.T) {
		// Sign authentic token with role: guard
		tokenStr, err := jwt.SignToken(userID, tenantID, "guard", secret)
		require.NoError(t, err)

		parts := strings.Split(tokenStr, ".")
		require.Len(t, parts, 3)

		// Decode payload and tamper role to super_admin
		rawPayload, err := base64.RawURLEncoding.DecodeString(parts[1])
		require.NoError(t, err)
		tamperedPayload := strings.Replace(string(rawPayload), `"role":"guard"`, `"role":"super_admin"`, 1)
		tamperedEncoded := base64.RawURLEncoding.EncodeToString([]byte(tamperedPayload))

		// Reconstruct token with tampered payload and original signature
		forgedToken := fmt.Sprintf("%s.%s.%s", parts[0], tamperedEncoded, parts[2])

		claims, err := jwt.ValidateToken(forgedToken, secret)
		assert.Error(t, err, "Payload tampering must invalidate token signature")
		assert.ErrorIs(t, err, jwt.ErrInvalidToken)
		assert.Nil(t, claims)
	})

	t.Run("ATTACK: Expired token rejection (immediate and distant)", func(t *testing.T) {
		// 10 seconds expired (exceeding 5s leeway)
		pastToken, err := jwt.SignTokenWithDuration(userID, tenantID, "guard", secret, -10*time.Second)
		require.NoError(t, err)

		claims, err := jwt.ValidateToken(pastToken, secret)
		assert.ErrorIs(t, err, jwt.ErrExpiredToken)
		assert.Nil(t, claims)

		// 30 days expired
		oldToken, err := jwt.SignTokenWithDuration(userID, tenantID, "guard", secret, -30*24*time.Hour)
		require.NoError(t, err)

		claims, err = jwt.ValidateToken(oldToken, secret)
		assert.ErrorIs(t, err, jwt.ErrExpiredToken)
		assert.Nil(t, claims)
	})

	t.Run("ATTACK: Short secret boundary enforcement (<32 bytes)", func(t *testing.T) {
		lengths := []int{0, 1, 8, 16, 31}
		for _, l := range lengths {
			shortSec := make([]byte, l)
			for i := range shortSec {
				shortSec[i] = 'a'
			}

			_, err := jwt.SignToken(userID, tenantID, "guard", shortSec)
			if l == 0 {
				assert.ErrorIs(t, err, jwt.ErrEmptySecret, "len=0 must return ErrEmptySecret")
			} else {
				assert.ErrorIs(t, err, jwt.ErrSecretTooShort, "len=%d must return ErrSecretTooShort", l)
			}

			_, err = jwt.ValidateToken("dummy.token.here", shortSec)
			if l == 0 {
				assert.ErrorIs(t, err, jwt.ErrEmptySecret)
			} else {
				assert.ErrorIs(t, err, jwt.ErrSecretTooShort)
			}
		}

		// Boundary test: exactly 32 bytes must be accepted
		exact32Sec := []byte("12345678901234567890123456789012")
		require.Len(t, exact32Sec, 32)
		tokenStr, err := jwt.SignToken(userID, tenantID, "guard", exact32Sec)
		assert.NoError(t, err, "32-byte secret must be accepted for signing")

		claims, err := jwt.ValidateToken(tokenStr, exact32Sec)
		assert.NoError(t, err, "32-byte secret must be accepted for validation")
		assert.NotNil(t, claims)
	})
}
