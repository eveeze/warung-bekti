package service

import (
	"context"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/repository"
	"github.com/google/uuid"
)

type RefillableService struct {
	db             *database.PostgresDB
	refillableRepo *repository.RefillableRepository
}

func NewRefillableService(db *database.PostgresDB, refillableRepo *repository.RefillableRepository) *RefillableService {
	return &RefillableService{
		db:             db,
		refillableRepo: refillableRepo,
	}
}

func (s *RefillableService) GetContainers(ctx context.Context) ([]domain.RefillableContainer, error) {
	return s.refillableRepo.GetContainers(ctx)
}

func (s *RefillableService) CreateContainer(ctx context.Context, input domain.RefillableContainer) (*domain.RefillableContainer, error) {
	if err := s.refillableRepo.Create(ctx, &input); err != nil {
		return nil, err
	}
	return &input, nil
}

func (s *RefillableService) GetContainerMovements(ctx context.Context, containerID string, page, perPage int) ([]domain.ContainerMovement, int64, error) {
	id, err := uuid.Parse(containerID)
	if err != nil {
		return nil, 0, domain.ErrInvalidInput
	}

	limit := perPage
	offset := (page - 1) * perPage

	return s.refillableRepo.GetMovements(ctx, id, limit, offset)
}

// AdjustStock allows manual adjustment
func (s *RefillableService) AdjustStock(ctx context.Context, input domain.ContainerMovement) (*domain.ContainerMovement, error) {
	// Transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Update Stock
	if err := s.refillableRepo.UpdateContainerStock(ctx, tx, input.ContainerID, input.EmptyChange, input.FullChange); err != nil {
		return nil, err
	}

	// Record Movement
	input.Type = domain.ContainerMovementAdjustment
	if err := s.refillableRepo.RecordMovement(ctx, tx, &input); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &input, nil
}
