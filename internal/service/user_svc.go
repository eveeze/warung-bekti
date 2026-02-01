package service

import (
	"context"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/pkg/password"
	"github.com/google/uuid"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context, p domain.UserListParams) ([]domain.User, int64, error) {
	if p.Page <= 0 { p.Page = 1 }
	if p.PerPage <= 0 { p.PerPage = 10 }
	return s.repo.List(ctx, p)
}

func (s *UserService) CreateUser(ctx context.Context, req domain.RegisterRequest) (*domain.User, error) {
	hash, err := password.Hash(req.Password)
	if err != nil { return nil, err }

	user := &domain.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         req.Role,
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, user); err != nil { return nil, err }
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, req domain.UpdateUserRequest) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil { return nil, err }

	user.Name = req.Name
	user.Email = req.Email
	user.Role = req.Role
	user.IsActive = req.IsActive

	if req.Password != "" {
		hash, err := password.Hash(req.Password)
		if err != nil { return nil, err }
		user.PasswordHash = hash
	}

	if err := s.repo.Update(ctx, user); err != nil { return nil, err }
	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}