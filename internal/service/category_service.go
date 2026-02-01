package service

import (
	"context"
	"errors"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/repository"
	"github.com/google/uuid"
)

type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

func (s *CategoryService) List(ctx context.Context, filter map[string]interface{}) ([]domain.CategoryResponse, error) {
	return s.categoryRepo.FindAll(ctx, filter)
}

func (s *CategoryService) Get(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	return s.categoryRepo.FindByID(ctx, id)
}

func (s *CategoryService) Create(ctx context.Context, input domain.CategoryCreateInput) (*domain.Category, error) {
	if input.Name == "" {
		return nil, errors.New("category name is required")
	}

	category := &domain.Category{
		Name:        input.Name,
		Description: input.Description,
		ParentID:    input.ParentID,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, input domain.CategoryUpdateInput) (*domain.Category, error) {
	existing, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("category not found")
	}

	if input.Name != nil {
		if *input.Name == "" {
			return nil, errors.New("category name cannot be empty")
		}
		existing.Name = *input.Name
	}
	if input.Description != nil {
		existing.Description = input.Description
	}
	if input.ParentID != nil {
		existing.ParentID = input.ParentID
	}
	if input.IsActive != nil {
		existing.IsActive = *input.IsActive
	}

	if err := s.categoryRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("category not found")
	}

	// Validation: Check if products exist
	hasProducts, err := s.categoryRepo.HasProducts(ctx, id)
	if err != nil {
		return err
	}
	if hasProducts {
		return errors.New("cannot delete category that contains active products")
	}

	return s.categoryRepo.Delete(ctx, id)
}
