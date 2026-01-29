package service

import (
	"context"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/repository"
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
