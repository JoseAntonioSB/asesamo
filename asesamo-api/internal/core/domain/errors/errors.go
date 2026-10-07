// Package errors defines standardized domain errors for ASESAMO,
// featuring UPPER_SNAKE_CASE error codes, HTTP status mapping,
// and user-facing messages in Spanish per DEVELOPMENT_RULES.md.
package errors

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
)

// AppError represents a structured domain error with error code,
// HTTP status code, and localized user-facing message.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error if wrapped.
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is reports whether any error in err's tree matches target.
func (e *AppError) Is(target error) bool {
	var t *AppError
	if targetErr, ok := target.(*AppError); ok {
		return e.Code == targetErr.Code
	}
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// IsNotFound returns true if the error indicates a record was not found.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, sql.ErrNoRows) {
		return true
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus == http.StatusNotFound
	}
	return false
}

// IsInvalidInput returns true if the error indicates invalid input.
func IsInvalidInput(err error) bool {
	if err == nil {
		return false
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus == http.StatusBadRequest || appErr.Code == "INVALID_INPUT"
	}
	return false
}

// IsTooManyRequests returns true if the error indicates rate limiting.
func IsTooManyRequests(err error) bool {
	if err == nil {
		return false
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus == http.StatusTooManyRequests || appErr.Code == "TOO_MANY_REQUESTS"
	}
	return false
}

// WithErr returns a copy of AppError wrapping an underlying cause.
func (e *AppError) WithErr(err error) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		HTTPStatus: e.HTTPStatus,
		Err:        err,
	}
}

// WithMessage returns a copy of AppError with a custom Spanish message.
func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    msg,
		HTTPStatus: e.HTTPStatus,
		Err:        e.Err,
	}
}

// Canonical domain errors.
var (
	ErrInvalidCredentials = &AppError{
		Code:       "INVALID_CREDENTIALS",
		Message:    "Credenciales inválidas. Verifique su correo y contraseña.",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrUserInactive = &AppError{
		Code:       "USER_INACTIVE",
		Message:    "La cuenta de usuario se encuentra inactiva.",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrUserNotFound = &AppError{
		Code:       "USER_NOT_FOUND",
		Message:    "Usuario no encontrado.",
		HTTPStatus: http.StatusNotFound,
	}
	ErrTenantNotFound = &AppError{
		Code:       "TENANT_NOT_FOUND",
		Message:    "Tenant no encontrado o no autorizado.",
		HTTPStatus: http.StatusNotFound,
	}
	ErrUnauthorized = &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "No autorizado. Se requiere token de autenticación válido.",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrForbidden = &AppError{
		Code:       "FORBIDDEN",
		Message:    "Acceso denegado. Permisos insuficientes.",
		HTTPStatus: http.StatusForbidden,
	}
	ErrInvalidToken = &AppError{
		Code:       "INVALID_TOKEN",
		Message:    "Token de autenticación inválido o corrupto.",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrExpiredToken = &AppError{
		Code:       "TOKEN_EXPIRED",
		Message:    "El token de autenticación ha vencido.",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrAccessPointNotFound = &AppError{
		Code:       "ACCESS_POINT_NOT_FOUND",
		Message:    "Punto de acceso no encontrado o no pertenece al tenant.",
		HTTPStatus: http.StatusNotFound,
	}
	ErrWorkerNotFound = &AppError{
		Code:       "WORKER_NOT_FOUND",
		Message:    "Trabajador no encontrado en el sistema.",
		HTTPStatus: http.StatusNotFound,
	}
	ErrContractorNotFound = &AppError{
		Code:       "CONTRACTOR_NOT_FOUND",
		Message:    "Empresa contratista no encontrada.",
		HTTPStatus: http.StatusNotFound,
	}
	ErrContractorNotActive = &AppError{
		Code:       "CONTRACTOR_NOT_ACTIVE",
		Message:    "El contratista no se encuentra activo o autorizado en esta planta.",
		HTTPStatus: http.StatusForbidden,
	}
	ErrDocumentExpired = &AppError{
		Code:       "DOCUMENT_EXPIRED",
		Message:    "Uno o más documentos requeridos han vencido.",
		HTTPStatus: http.StatusUnprocessableEntity,
	}
	ErrDocumentMissing = &AppError{
		Code:       "DOCUMENT_MISSING",
		Message:    "Faltan documentos obligatorios para el ingreso.",
		HTTPStatus: http.StatusUnprocessableEntity,
	}
	ErrInvalidInput = &AppError{
		Code:       "INVALID_INPUT",
		Message:    "Los datos proporcionados son inválidos o faltan campos obligatorios.",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrInternal = &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "Ha ocurrido un error interno del servidor.",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrDatabase = &AppError{
		Code:       "DATABASE_ERROR",
		Message:    "Error de comunicación con la base de datos.",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrTooManyRequests = &AppError{
		Code:       "TOO_MANY_REQUESTS",
		Message:    "Demasiadas solicitudes. Intente más tarde.",
		HTTPStatus: http.StatusTooManyRequests,
	}
)
