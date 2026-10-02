package errors_test

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"

	domainErrors "asesamo-api/internal/core/domain/errors"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	err := domainErrors.ErrInvalidCredentials
	assert.Equal(t, "INVALID_CREDENTIALS", err.Code)
	assert.Equal(t, http.StatusUnauthorized, err.HTTPStatus)
	assert.Contains(t, err.Error(), "[INVALID_CREDENTIALS]")

	cause := errors.New("bcrypt mismatch")
	wrapped := err.WithErr(cause)
	assert.ErrorIs(t, wrapped, cause)
	assert.Equal(t, cause, wrapped.Unwrap())
	assert.Contains(t, wrapped.Error(), "bcrypt mismatch")

	withMsg := err.WithMessage("Custom Spanish error")
	assert.Equal(t, "Custom Spanish error", withMsg.Message)
	assert.Equal(t, err.Code, withMsg.Code)

	// Test Is method
	assert.True(t, wrapped.Is(domainErrors.ErrInvalidCredentials))
	assert.True(t, wrapped.Is(err))
	assert.False(t, wrapped.Is(domainErrors.ErrUserInactive))
	assert.False(t, wrapped.Is(errors.New("generic error")))
}

func TestIsNotFound(t *testing.T) {
	assert.False(t, domainErrors.IsNotFound(nil))
	assert.True(t, domainErrors.IsNotFound(sql.ErrNoRows))
	assert.True(t, domainErrors.IsNotFound(domainErrors.ErrUserNotFound))
	assert.True(t, domainErrors.IsNotFound(domainErrors.ErrTenantNotFound))
	assert.True(t, domainErrors.IsNotFound(domainErrors.ErrAccessPointNotFound))

	customNotFound := &domainErrors.AppError{
		Code:       "CUSTOM_NOT_FOUND",
		Message:    "No encontrado",
		HTTPStatus: http.StatusNotFound,
	}
	assert.True(t, domainErrors.IsNotFound(customNotFound))

	assert.False(t, domainErrors.IsNotFound(domainErrors.ErrForbidden))
	assert.False(t, domainErrors.IsNotFound(errors.New("other error")))
}
