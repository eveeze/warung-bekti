package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/middleware"
	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/pkg/validator"
	"github.com/eveeze/warung-backend/internal/service"
)

// PaymentHandler handles payment HTTP requests
type PaymentHandler struct {
	paymentSvc *service.PaymentService
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(paymentSvc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentSvc: paymentSvc}
}

// GenerateSnapToken generates a Snap token for payment
// POST /payments/snap
func (h *PaymentHandler) GenerateSnapToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TransactionID string  `json:"transaction_id"`
		CustomerName  *string `json:"customer_name,omitempty"`
		CustomerEmail *string `json:"customer_email,omitempty"`
		CustomerPhone *string `json:"customer_phone,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	v := validator.New()
	v.Required("transaction_id", req.TransactionID, "Transaction ID is required")

	if v.HasErrors() {
		response.ValidationError(w, v.Errors())
		return
	}

	transactionID, err := uuid.Parse(req.TransactionID)
	if err != nil {
		response.BadRequest(w, "Invalid transaction ID format")
		return
	}

	snapReq := domain.SnapTokenRequest{
		TransactionID: transactionID,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone,
	}

	result, err := h.paymentSvc.GenerateSnapToken(r.Context(), snapReq)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, "Snap token generated successfully", result)
}

// HandleNotification handles Midtrans webhook notification
// POST /payments/notification (PUBLIC - no auth required)
func (h *PaymentHandler) HandleNotification(w http.ResponseWriter, r *http.Request) {
	var notification domain.MidtransNotification

	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		response.BadRequest(w, "Invalid notification payload")
		return
	}

	if err := h.paymentSvc.HandleNotification(r.Context(), notification); err != nil {
		// Log error but return OK to Midtrans to prevent retry storm
		// In production, you'd want to log this properly
		response.OK(w, "Notification received", nil)
		return
	}

	response.OK(w, "Notification processed", nil)
}

// ManualVerify manually verifies a payment
// POST /payments/{id}/manual-verify
func (h *PaymentHandler) ManualVerify(w http.ResponseWriter, r *http.Request) {
	paymentIDStr := r.PathValue("id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid payment ID format")
		return
	}

	// Get user from context for audit
	claims := middleware.GetUserFromContext(r.Context())
	verifiedBy := "system"
	if claims != nil {
		verifiedBy = claims.Username
	}

	if err := h.paymentSvc.ManualVerify(r.Context(), paymentID, verifiedBy); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, "Payment verified manually", nil)
}

// GetPaymentByTransaction gets payment info for a transaction
// GET /payments/transaction/{id}
func (h *PaymentHandler) GetPaymentByTransaction(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.PathValue("id")
	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid transaction ID format")
		return
	}

	payment, err := h.paymentSvc.GetPaymentByTransactionID(r.Context(), transactionID)
	if err != nil {
		if err == domain.ErrNotFound {
			response.NotFound(w, "Payment record not found")
			return
		}
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, "Payment record retrieved", payment)
}
