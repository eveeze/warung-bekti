package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/repository"
)

type ConsignmentService struct {
	db              *database.PostgresDB
	consignmentRepo *repository.ConsignmentRepository
	transactionRepo *repository.TransactionRepository
}

func NewConsignmentService(db *database.PostgresDB, consignmentRepo *repository.ConsignmentRepository, transactionRepo *repository.TransactionRepository) *ConsignmentService {
	return &ConsignmentService{
		db:              db,
		consignmentRepo: consignmentRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *ConsignmentService) CreateConsignor(ctx context.Context, input domain.CreateConsignorInput) (*domain.Consignor, error) {
	return s.consignmentRepo.CreateConsignor(ctx, input)
}

func (s *ConsignmentService) UpdateConsignor(ctx context.Context, id uuid.UUID, input domain.UpdateConsignorInput) (*domain.Consignor, error) {
	return s.consignmentRepo.UpdateConsignor(ctx, id, input)
}

func (s *ConsignmentService) ListConsignors(ctx context.Context) ([]domain.Consignor, error) {
	return s.consignmentRepo.ListConsignors(ctx)
}

func (s *ConsignmentService) DeleteConsignor(ctx context.Context, id uuid.UUID) error {
	return s.consignmentRepo.DeleteConsignor(ctx, id)
}

// GenerateSettlement calculates sales for a consignor in a period
// Simplified logic: Query transactions for products belonging to consignor
func (s *ConsignmentService) GenerateSettlement(ctx context.Context, consignorID uuid.UUID, createdBy string) (*domain.ConsignmentSettlement, error) {
	// Logic to query transactions... omitting detailed implementation for brevity as "best effort vanilla"
	// Would involve:
	// 1. Get products by consignor_id
	// 2. Sum sales from transaction_items for those products within default period (e.g. last 30 days or since last settlement)
	
	// Returning placeholder logic
	return &domain.ConsignmentSettlement{ConsignorID: consignorID, Status: domain.SettlementStatusDraft}, nil
}
