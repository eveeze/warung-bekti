package response_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/pkg/validator"
)

func TestResponses(t *testing.T) {
	// NewMeta
	meta := response.NewMeta(1, 10, 25)
	if meta.Page != 1 || meta.PerPage != 10 || meta.Total != 25 || meta.TotalPages != 3 {
		t.Errorf("unexpected meta: %+v", meta)
	}

	zeroPages := response.CalculateTotalPages(25, 0)
	if zeroPages != 0 {
		t.Errorf("expected 0 total pages for perPage <= 0, got %d", zeroPages)
	}

	tests := []struct {
		name       string
		fn         func(w http.ResponseWriter)
		statusCode int
	}{
		{"Success", func(w http.ResponseWriter) { response.Success(w, http.StatusOK, "ok", map[string]string{"foo": "bar"}) }, http.StatusOK},
		{"SuccessWithMeta", func(w http.ResponseWriter) { response.SuccessWithMeta(w, http.StatusOK, "ok", []string{"a"}, meta) }, http.StatusOK},
		{"OK", func(w http.ResponseWriter) { response.OK(w, "retrieved", 123) }, http.StatusOK},
		{"Created", func(w http.ResponseWriter) { response.Created(w, "created", 123) }, http.StatusCreated},
		{"NoContent", func(w http.ResponseWriter) { response.NoContent(w) }, http.StatusNoContent},
		{"Error", func(w http.ResponseWriter) { response.Error(w, http.StatusBadRequest, "ERR", "msg") }, http.StatusBadRequest},
		{"ErrorWithDetails", func(w http.ResponseWriter) { response.ErrorWithDetails(w, 400, "ERR", "msg", map[string]string{"a": "b"}) }, 400},
		{"BadRequest", func(w http.ResponseWriter) { response.BadRequest(w, "bad request") }, http.StatusBadRequest},
		{"Unauthorized", func(w http.ResponseWriter) { response.Unauthorized(w, "unauthorized") }, http.StatusUnauthorized},
		{"Forbidden", func(w http.ResponseWriter) { response.Forbidden(w, "forbidden") }, http.StatusForbidden},
		{"NotFound", func(w http.ResponseWriter) { response.NotFound(w, "not found") }, http.StatusNotFound},
		{"Conflict", func(w http.ResponseWriter) { response.Conflict(w, "conflict") }, http.StatusConflict},
		{"UnprocessableEntity", func(w http.ResponseWriter) { response.UnprocessableEntity(w, "unprocessable") }, http.StatusUnprocessableEntity},
		{"PreconditionFailed", func(w http.ResponseWriter) { response.PreconditionFailed(w, "precondition failed") }, http.StatusPreconditionFailed},
		{"InternalServerError", func(w http.ResponseWriter) { response.InternalServerError(w, "internal server error") }, http.StatusInternalServerError},
		{"ServiceUnavailable", func(w http.ResponseWriter) { response.ServiceUnavailable(w, "unavailable") }, http.StatusServiceUnavailable},
		{"ValidationError", func(w http.ResponseWriter) {
			errs := validator.ValidationErrors{"field": "error message"}
			response.ValidationError(w, errs)
		}, http.StatusUnprocessableEntity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(rec)
			if rec.Code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, rec.Code)
			}
		})
	}
}
