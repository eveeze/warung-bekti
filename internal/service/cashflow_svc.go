package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/repository"
)

type CashFlowService struct {
	db           *database.PostgresDB
	cashFlowRepo *repository.CashFlowRepository
}

func NewCashFlowService(db *database.PostgresDB, cashFlowRepo *repository.CashFlowRepository) *CashFlowService {
	return &CashFlowService{
		db:           db,
		cashFlowRepo: cashFlowRepo,
	}
}

func (s *CashFlowService) GetCategories(ctx context.Context) ([]domain.CashFlowCategory, error) {
	return s.cashFlowRepo.GetCategories(ctx)
}

func (s *CashFlowService) OpenDrawer(ctx context.Context, input domain.OpenDrawerInput) (*domain.CashDrawerSession, error) {
	// Check validations if needed
	if input.OpeningBalance < 0 {
		return nil, fmt.Errorf("opening balance cannot be negative")
	}
	return s.cashFlowRepo.OpenDrawer(ctx, input)
}

func (s *CashFlowService) CloseDrawer(ctx context.Context, input domain.CloseDrawerInput) (*domain.CashDrawerSession, error) {
	if input.ClosingBalance < 0 {
		return nil, fmt.Errorf("closing balance cannot be negative")
	}
	
	// Verify session exists and is open
	session, err := s.cashFlowRepo.GetCurrentSession(ctx)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, fmt.Errorf("no open session found")
		}
		return nil, err
	}
	
	if session.ID != input.SessionID {
		return nil, fmt.Errorf("session ID mismatch or already closed")
	}

	return s.cashFlowRepo.CloseDrawer(ctx, input)
}

func (s *CashFlowService) GetCurrentSession(ctx context.Context) (*domain.CashDrawerSession, error) {
	return s.cashFlowRepo.GetCurrentSession(ctx)
}

func (s *CashFlowService) RecordCashFlow(ctx context.Context, input domain.CashFlowInput) (*domain.CashFlowRecord, error) {
	if input.Amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	// Attach to current session if open
	session, err := s.cashFlowRepo.GetCurrentSession(ctx)
	if err != nil && err != domain.ErrNotFound {
		return nil, err
	}
	
	// If no session open, we still record it but without session_id (or error? Warung usually needs open drawer)
	// Let's allow without session for now, or maybe require it. 
	// Ideally, all cash moves happen within a session.
	// But expenses might happen before opening? Let's keep it optional but recommended.
	
	var sessionID *uuid.UUID
	if session != nil {
		sessionID = &session.ID
	}

	return s.cashFlowRepo.RecordCashFlow(ctx, nil, input, sessionID, nil, nil)
}

func (s *CashFlowService) ListCashFlows(ctx context.Context, filter domain.CashFlowFilter) ([]domain.CashFlowRecord, int64, error) {
	return s.cashFlowRepo.ListCashFlows(ctx, filter)
}
