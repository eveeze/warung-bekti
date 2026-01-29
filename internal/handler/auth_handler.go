package handler

import (
	"encoding/json"
	"net/http"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/pkg/validator"
	"github.com/eveeze/warung-backend/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
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
	v.Required("role", string(req.Role), "role is required")
	v.InSlice("role", string(req.Role), []string{"cashier", "inventory"}, "invalid role: admin registration is restricted")

	if v.HasErrors() {
		response.ValidationError(w, v.Errors())
		return
	}

	resp, err := h.authSvc.Register(r.Context(), req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, "User registered successfully", resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	v := validator.New()
	v.Required("email", req.Email, "email is required")
	v.Email("email", req.Email, "invalid email format")
	v.Required("password", req.Password, "password is required")

	if v.HasErrors() {
		response.ValidationError(w, v.Errors())
		return
	}

	resp, err := h.authSvc.Login(r.Context(), req)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}

	response.OK(w, "Login successful", resp)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		response.BadRequest(w, "Refresh token is required")
		return
	}

	resp, err := h.authSvc.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}

	response.OK(w, "Token refreshed successfully", resp)
}
