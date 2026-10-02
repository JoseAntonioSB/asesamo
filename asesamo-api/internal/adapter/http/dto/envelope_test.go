package dto_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"asesamo-api/internal/adapter/http/dto"
	domainErrors "asesamo-api/internal/core/domain/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"foo": "bar"}

	dto.WriteSuccess(rec, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp dto.SuccessResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	respMap, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "bar", respMap["foo"])
}

func TestWriteError_AppError(t *testing.T) {
	rec := httptest.NewRecorder()
	appErr := domainErrors.ErrInvalidCredentials

	dto.WriteError(rec, appErr)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	var resp dto.ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "INVALID_CREDENTIALS", resp.Error.Code)
	assert.Equal(t, appErr.Message, resp.Error.Message)
}

func TestWriteError_GenericError(t *testing.T) {
	rec := httptest.NewRecorder()
	genericErr := errors.New("something unexpected happened")

	dto.WriteError(rec, genericErr)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp dto.ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "INTERNAL_ERROR", resp.Error.Code)
	assert.Equal(t, "Ha ocurrido un error interno del servidor.", resp.Error.Message)
}
