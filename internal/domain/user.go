package domain

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleCashier   UserRole = "cashier"
	RoleInventory UserRole = "inventory"
)

type UserClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type User struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         UserRole   `json:"role"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}

type UserListParams struct {
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Search  string `json:"search"`
}

type UpdateUserRequest struct {
	Name     *string   `json:"name"`
	Email    *string   `json:"email"`
	Password *string   `json:"password"`
	Role     *UserRole `json:"role"`
	IsActive *bool     `json:"is_active"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, params UserListParams) ([]User, int64, error) // Pakai int64
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Name     string   `json:"name" validate:"required"`
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password" validate:"required,min=6"`
	Role     UserRole `json:"role" validate:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}