package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"asesamo-api/internal/adapter/http/handler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockPinger struct {
	mock.Mock
}

func (m *MockPinger) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestHealthHandler(t *testing.T) {
	t.Run("returns 200 OK when database is connected", func(t *testing.T) {
		mockPinger := new(MockPinger)
		mockPinger.On("Ping", mock.Anything).Return(nil)

		h := handler.NewHealthHandler(mockPinger, "1.0.0")

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		h.Check(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var resp handler.HealthResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "ok", resp.Status)
		assert.Equal(t, "connected", resp.Database)
		assert.Equal(t, "1.0.0", resp.Version)
		assert.NotEmpty(t, resp.Timestamp)

		mockPinger.AssertExpectations(t)
	})

	t.Run("returns 503 Service Unavailable when database ping fails", func(t *testing.T) {
		mockPinger := new(MockPinger)
		mockPinger.On("Ping", mock.Anything).Return(errors.New("connection refused"))

		h := handler.NewHealthHandler(mockPinger, "1.0.0")

		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		rec := httptest.NewRecorder()

		h.Check(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

		var resp handler.HealthResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "degraded", resp.Status)
		assert.Equal(t, "disconnected", resp.Database)
		assert.Equal(t, "1.0.0", resp.Version)

		mockPinger.AssertExpectations(t)
	})

	t.Run("nil pinger defaults to connected", func(t *testing.T) {
		h := handler.NewHealthHandler(nil, "")

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		h.Check(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp handler.HealthResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "ok", resp.Status)
		assert.Equal(t, "connected", resp.Database)
		assert.Equal(t, "1.0.0", resp.Version)
	})
}
