package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/pkg/validator"
	"github.com/eveeze/warung-backend/internal/service"
	"github.com/google/uuid"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	v := validator.New()
	v.Required("name", req.Name, "name is required")
	v.Required("email", req.Email, "email is required")
	v.Email("email", req.Email, "invalid email format")
	v.Required("password", req.Password, "password is required")
	v.MinLength("password", req.Password, 6, "password must be at least 6 characters")

	if v.HasErrors() {
		response.ValidationError(w, v.Errors())
		return
	}

	user, err := h.userSvc.CreateUser(r.Context(), req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, "User created successfully", user)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page <= 0 { page = 1 }
	if perPage <= 0 { perPage = 10 }

	params := domain.UserListParams{
		Page:    page,
		PerPage: perPage,
		Search:  r.URL.Query().Get("search"),
	}

	users, total, err := h.userSvc.ListUsers(r.Context(), params)
	if err != nil {
		response.InternalServerError(w, "Failed to list users")
		return
	}

	meta := response.NewMeta(params.Page, params.PerPage, total)
	response.SuccessWithMeta(w, http.StatusOK, "Users retrieved", users, meta)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	user, err := h.userSvc.GetByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	response.OK(w, "User retrieved", user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	var req domain.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	v := validator.New()
	v.Required("name", req.Name, "name is required")
	v.Email("email", req.Email, "invalid email")

	if v.HasErrors() {
		response.ValidationError(w, v.Errors())
		return
	}

	user, err := h.userSvc.UpdateUser(r.Context(), id, req)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, "User updated successfully", user)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	if err := h.userSvc.DeleteUser(r.Context(), id); err != nil {
		response.InternalServerError(w, "Failed to delete user")
		return
	}

	response.OK(w, "User deleted successfully", nil)
}