// Package jwt provides stateless HMAC-SHA256 JWT token generation,
// parsing, and claims validation for ASESAMO.
package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Standard duration constants per DEVELOPMENT_RULES.md line 195.
const (
	TokenDuration = 1 * time.Hour
	MinSecretLen  = 32
)

// Sentinel package errors.
var (
	ErrInvalidToken   = errors.New("jwt: token is invalid or corrupted")
	ErrExpiredToken   = errors.New("jwt: token has expired")
	ErrEmptySecret    = errors.New("jwt: secret cannot be empty")
	ErrSecretTooShort = errors.New("jwt: secret must be at least 32 characters long")
	ErrInvalidSubject = errors.New("jwt: token subject is not a valid UUID")
)

// Claims represents the standard and custom JWT claims for ASESAMO.
type Claims struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Role     string    `json:"role"`
	Email    string    `json:"email,omitempty"`
	jwtlib.RegisteredClaims
}

// UserID returns the parsed user UUID from the Subject ('sub') claim.
func (c *Claims) UserID() (uuid.UUID, error) {
	if c.Subject == "" {
		return uuid.Nil, ErrInvalidSubject
	}
	id, err := uuid.Parse(c.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %v", ErrInvalidSubject, err)
	}
	return id, nil
}

// SignToken generates and signs an HS256 JWT for a given user, tenant, and role.
// It sets 'exp' to 1 hour from issuance.
func SignToken(userID uuid.UUID, tenantID uuid.UUID, role string, secret []byte) (string, error) {
	return SignTokenWithDuration(userID, tenantID, role, secret, TokenDuration)
}

// SignTokenWithDuration generates and signs an HS256 JWT with a custom lifetime.
func SignTokenWithDuration(userID uuid.UUID, tenantID uuid.UUID, role string, secret []byte, duration time.Duration) (string, error) {
	if len(secret) == 0 {
		return "", ErrEmptySecret
	}
	if len(secret) < MinSecretLen {
		return "", ErrSecretTooShort
	}

	now := time.Now().UTC()
	claims := Claims{
		TenantID: tenantID,
		Role:     role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwtlib.NewNumericDate(now),
			NotBefore: jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(duration)),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("jwt: signing failed: %w", err)
	}

	return signedToken, nil
}

// ValidateToken parses, verifies the signature, and validates claims for a token string.
func ValidateToken(tokenStr string, secret []byte) (*Claims, error) {
	if len(secret) == 0 {
		return nil, ErrEmptySecret
	}
	if len(secret) < MinSecretLen {
		return nil, ErrSecretTooShort
	}

	token, err := jwtlib.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(t *jwtlib.Token) (interface{}, error) {
			// Enforce HMAC-SHA256 explicitly to prevent algorithm confusion attacks
			if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("%w: unexpected signing method %v", ErrInvalidToken, t.Header["alg"])
			}
			return secret, nil
		},
		jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Name}),
		jwtlib.WithLeeway(5*time.Second),
	)

	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
