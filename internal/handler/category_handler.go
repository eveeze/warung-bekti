package handler

import (
	"encoding/json"
	"net/http"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/service"
	"github.com/google/uuid"
)

type CategoryHandler struct {
	categorySvc *service.CategoryService
}

func NewCategoryHandler(categorySvc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categorySvc: categorySvc}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := make(map[string]interface{})
	search := r.URL.Query().Get("search")
	if search != "" {
		filter["search"] = search
	}

	categories, err := h.categorySvc.List(r.Context(), filter)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}
	response.OK(w, "Categories retrieved", categories)
}

func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid ID format")
		return
	}

	category, err := h.categorySvc.Get(r.Context(), id)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}
	if category == nil {
		response.NotFound(w, "Category not found")
		return
	}

	response.OK(w, "Category details", category)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input domain.CategoryCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	category, err := h.categorySvc.Create(r.Context(), input)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}

	response.Created(w, "Category created successfully", category)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid ID format")
		return
	}

	var input domain.CategoryUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	category, err := h.categorySvc.Update(r.Context(), id, input)
	if err != nil {
		if err.Error() == "category not found" {
			response.NotFound(w, err.Error())
			return
		}
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, "Category updated successfully", category)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid ID format")
		return
	}

	err = h.categorySvc.Delete(r.Context(), id)
	if err != nil {
		if err.Error() == "category not found" {
			response.NotFound(w, err.Error())
			return
		}
		if err.Error() == "cannot delete category that contains active products" {
			response.BadRequest(w, err.Error())
			return
		}
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, "Category deleted successfully", nil)
}
