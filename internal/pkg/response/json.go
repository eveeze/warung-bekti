package response

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorInfo represents error details in a response
type ErrorInfo struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Meta represents pagination metadata
type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// JSON sends a JSON response with the given status code
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, `{"success":false,"error":{"code":"INTERNAL_ERROR","message":"Failed to encode response"}}`, http.StatusInternalServerError)
		}
	}
}

// Success sends a successful response
func Success(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	JSON(w, statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithMeta sends a successful response with pagination metadata
func SuccessWithMeta(w http.ResponseWriter, statusCode int, message string, data interface{}, meta *Meta) {
	JSON(w, statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Error sends an error response
func Error(w http.ResponseWriter, statusCode int, code, message string) {
	JSON(w, statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

// ErrorWithDetails sends an error response with additional details
func ErrorWithDetails(w http.ResponseWriter, statusCode int, code, message string, details map[string]string) {
	JSON(w, statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// Common error responses

// BadRequest sends a 400 Bad Request response
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound sends a 404 Not Found response
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict sends a 409 Conflict response
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, "CONFLICT", message)
}

// UnprocessableEntity sends a 422 Unprocessable Entity response
func UnprocessableEntity(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", message)
}

// ValidationError sends a 422 response with validation details
func ValidationError(w http.ResponseWriter, details map[string]string) {
	ErrorWithDetails(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validation failed", details)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// ServiceUnavailable sends a 503 Service Unavailable response
func ServiceUnavailable(w http.ResponseWriter, message string) {
	Error(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message)
}

// Created sends a 201 Created response
func Created(w http.ResponseWriter, message string, data interface{}) {
	Success(w, http.StatusCreated, message, data)
}

// OK sends a 200 OK response
func OK(w http.ResponseWriter, message string, data interface{}) {
	Success(w, http.StatusOK, message, data)
}

// NoContent sends a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// CalculateTotalPages calculates the total number of pages
func CalculateTotalPages(total int64, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	pages := int(total) / perPage
	if int(total)%perPage > 0 {
		pages++
	}
	return pages
}

// NewMeta creates a new Meta object with calculated total pages
func NewMeta(page, perPage int, total int64) *Meta {
	return &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: CalculateTotalPages(total, perPage),
	}
}
