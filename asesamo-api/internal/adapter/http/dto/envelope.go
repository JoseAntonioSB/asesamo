// Package dto contains request and response Data Transfer Objects for HTTP adapters.
package dto

import (
	"encoding/json"
	"errors"
	"net/http"

	domainErrors "asesamo-api/internal/core/domain/errors"
)

// SuccessResponse defines the standard success envelope.
type SuccessResponse struct {
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta,omitempty"`
}

// ErrorDetail contains the code and message in an error envelope.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse defines the standard error envelope.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// WriteJSON serializes payload as JSON and writes the specified HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

// PaginationMeta represents pagination metadata for list responses.
type PaginationMeta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

// WriteSuccess wraps data in the standard success envelope and writes HTTP 200 OK.
func WriteSuccess(w http.ResponseWriter, status int, data interface{}) {
	WriteJSON(w, status, SuccessResponse{Data: data})
}

// WriteSuccessWithMeta wraps data and pagination meta in the standard success envelope.
func WriteSuccessWithMeta(w http.ResponseWriter, status int, data interface{}, meta interface{}) {
	WriteJSON(w, status, SuccessResponse{Data: data, Meta: meta})
}

// WriteError inspects err and writes the standard error envelope with the appropriate HTTP status code.
func WriteError(w http.ResponseWriter, err error) {
	var appErr *domainErrors.AppError
	if errors.As(err, &appErr) {
		WriteJSON(w, appErr.HTTPStatus, ErrorResponse{
			Error: ErrorDetail{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
		})
		return
	}

	// Fallback to internal error
	WriteJSON(w, http.StatusInternalServerError, ErrorResponse{
		Error: ErrorDetail{
			Code:    "INTERNAL_ERROR",
			Message: "Ha ocurrido un error interno del servidor.",
		},
	})
}
