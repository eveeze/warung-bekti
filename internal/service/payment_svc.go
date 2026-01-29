package service

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/repository"
)

// PaymentService handles payment business logic
type PaymentService struct {
	db              *database.PostgresDB
	paymentRepo     *repository.PaymentRepository
	transactionRepo *repository.TransactionRepository
	cfg             *config.MidtransConfig
	httpClient      *http.Client
}

// NewPaymentService creates a new PaymentService
func NewPaymentService(
	db *database.PostgresDB,
	paymentRepo *repository.PaymentRepository,
	transactionRepo *repository.TransactionRepository,
	cfg *config.MidtransConfig,
) *PaymentService {
	return &PaymentService{
		db:              db,
		paymentRepo:     paymentRepo,
		transactionRepo: transactionRepo,
		cfg:             cfg,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
	}
}

// getSnapURL returns the Midtrans Snap API URL based on environment
func (s *PaymentService) getSnapURL() string {
	if s.cfg.Environment == "production" {
		return "https://app.midtrans.com/snap/v1/transactions"
	}
	return "https://app.sandbox.midtrans.com/snap/v1/transactions"
}

// GenerateSnapToken generates a Snap token for QRIS/payment
func (s *PaymentService) GenerateSnapToken(ctx context.Context, req domain.SnapTokenRequest) (*domain.SnapTokenResponse, error) {
	// Get transaction details
	transaction, err := s.transactionRepo.GetByID(ctx, req.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	// Generate unique order ID
	orderID := fmt.Sprintf("TRX-%s-%d", transaction.ID.String()[:8], time.Now().Unix())

	// Create payment record first
	paymentRecord := &domain.PaymentRecord{
		TransactionID: req.TransactionID,
		OrderID:       orderID,
		GrossAmount:   req.GrossAmount,
		Currency:      "IDR",
		Status:        domain.PaymentStatusPending,
	}

	if err := s.paymentRepo.Create(ctx, nil, paymentRecord); err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// Build Midtrans Snap request
	snapRequest := map[string]interface{}{
		"transaction_details": map[string]interface{}{
			"order_id":     orderID,
			"gross_amount": req.GrossAmount,
		},
	}

	// Add customer details if provided
	if req.CustomerName != nil || req.CustomerEmail != nil || req.CustomerPhone != nil {
		customerDetails := make(map[string]interface{})
		if req.CustomerName != nil {
			customerDetails["first_name"] = *req.CustomerName
		}
		if req.CustomerEmail != nil {
			customerDetails["email"] = *req.CustomerEmail
		}
		if req.CustomerPhone != nil {
			customerDetails["phone"] = *req.CustomerPhone
		}
		snapRequest["customer_details"] = customerDetails
	}

	// Add item details if provided
	if len(req.ItemDetails) > 0 {
		items := make([]map[string]interface{}, len(req.ItemDetails))
		for i, item := range req.ItemDetails {
			items[i] = map[string]interface{}{
				"id":       item.ID,
				"name":     item.Name,
				"price":    item.Price,
				"quantity": item.Quantity,
			}
		}
		snapRequest["item_details"] = items
	}

	// Enable QRIS and other payment methods
	snapRequest["enabled_payments"] = []string{"qris", "gopay", "shopeepay", "other_qris"}

	// Call Midtrans API
	body, err := json.Marshal(snapRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.getSnapURL(), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.SetBasicAuth(s.cfg.ServerKey, "")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Midtrans: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("midtrans error: %s", string(respBody))
	}

	var snapResp struct {
		Token       string `json:"token"`
		RedirectURL string `json:"redirect_url"`
	}
	if err := json.Unmarshal(respBody, &snapResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Update payment record with token
	if err := s.paymentRepo.UpdateSnapToken(ctx, paymentRecord.ID, snapResp.Token, snapResp.RedirectURL); err != nil {
		return nil, fmt.Errorf("failed to update snap token: %w", err)
	}

	return &domain.SnapTokenResponse{
		Token:       snapResp.Token,
		RedirectURL: snapResp.RedirectURL,
		OrderID:     orderID,
	}, nil
}

// HandleNotification processes Midtrans webhook notification
func (s *PaymentService) HandleNotification(ctx context.Context, notification domain.MidtransNotification) error {
	// Verify signature
	if !s.verifySignature(notification) {
		return fmt.Errorf("invalid signature")
	}

	// Get payment record
	paymentRecord, err := s.paymentRepo.GetByOrderID(ctx, notification.OrderID)
	if err != nil {
		return fmt.Errorf("payment record not found: %w", err)
	}

	// Map Midtrans status to our status
	var status domain.PaymentStatus
	switch notification.TransactionStatus {
	case "capture":
		if notification.FraudStatus == "accept" {
			status = domain.PaymentStatusCapture
		} else {
			status = domain.PaymentStatusPending
		}
	case "settlement":
		status = domain.PaymentStatusSettlement
	case "pending":
		status = domain.PaymentStatusPending
	case "deny":
		status = domain.PaymentStatusDeny
	case "cancel":
		status = domain.PaymentStatusCancel
	case "expire":
		status = domain.PaymentStatusExpire
	case "failure":
		status = domain.PaymentStatusFailure
	case "refund":
		status = domain.PaymentStatusRefund
	case "partial_refund":
		status = domain.PaymentStatusPartialRefund
	default:
		status = domain.PaymentStatusPending
	}

	// Store notification as JSON
	notifJSON, _ := json.Marshal(notification)

	// Update payment status
	if err := s.paymentRepo.UpdateStatus(ctx, paymentRecord.ID, status, notifJSON); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Update transaction status if payment is successful
	if status.IsSuccess() {
		if err := s.transactionRepo.UpdateStatus(ctx, paymentRecord.TransactionID, domain.TransactionStatusCompleted); err != nil {
			return fmt.Errorf("failed to update transaction status: %w", err)
		}
	}

	return nil
}

// verifySignature verifies the Midtrans notification signature
func (s *PaymentService) verifySignature(notification domain.MidtransNotification) bool {
	// Signature = SHA512(order_id + status_code + gross_amount + server_key)
	data := notification.OrderID + notification.StatusCode + notification.GrossAmount + s.cfg.ServerKey
	hash := sha512.Sum512([]byte(data))
	calculatedSignature := hex.EncodeToString(hash[:])

	return calculatedSignature == notification.SignatureKey
}

// ManualVerify manually marks a payment as successful
func (s *PaymentService) ManualVerify(ctx context.Context, paymentID uuid.UUID, verifiedBy string) error {
	paymentRecord, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("payment record not found: %w", err)
	}

	if paymentRecord.Status.IsFinal() && paymentRecord.Status.IsSuccess() {
		return fmt.Errorf("payment already verified")
	}

	// Create manual verification response
	manualResp := map[string]interface{}{
		"manual_verification": true,
		"verified_by":         verifiedBy,
		"verified_at":         time.Now().Format(time.RFC3339),
	}
	respJSON, _ := json.Marshal(manualResp)

	// Update payment status
	if err := s.paymentRepo.UpdateStatus(ctx, paymentRecord.ID, domain.PaymentStatusSettlement, respJSON); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Update transaction status
	if err := s.transactionRepo.UpdateStatus(ctx, paymentRecord.TransactionID, domain.TransactionStatusCompleted); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

// GetPaymentByTransactionID retrieves payment info for a transaction
func (s *PaymentService) GetPaymentByTransactionID(ctx context.Context, transactionID uuid.UUID) (*domain.PaymentRecord, error) {
	return s.paymentRepo.GetByTransactionID(ctx, transactionID)
}
