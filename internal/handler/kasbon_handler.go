package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/pkg/pdf"
	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/pkg/validator"
	"github.com/eveeze/warung-backend/internal/repository"
)

// KasbonHandler handles kasbon endpoints
type KasbonHandler struct {
	kasbonRepo   *repository.KasbonRepository
	customerRepo *repository.CustomerRepository
}

// NewKasbonHandler creates a new KasbonHandler
func NewKasbonHandler(kasbonRepo *repository.KasbonRepository, customerRepo *repository.CustomerRepository) *KasbonHandler {
	return &KasbonHandler{kasbonRepo: kasbonRepo, customerRepo: customerRepo}
}

// GetHistory retrieves kasbon history for a customer
func (h *KasbonHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid customer ID")
		return
	}

	query := r.URL.Query()
	filter := domain.KasbonFilter{CustomerID: &customerID, Page: 1, PerPage: 20}

	if page := query.Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filter.Page = p
		}
	}
	if perPage := query.Get("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil {
			filter.PerPage = pp
		}
	}
	if typeStr := query.Get("type"); typeStr != "" {
		t := domain.KasbonType(typeStr)
		filter.Type = &t
	}

	records, total, err := h.kasbonRepo.GetByCustomer(r.Context(), customerID, filter)
	if err != nil {
		response.InternalServerError(w, "Failed to get kasbon history")
		return
	}

	meta := response.NewMeta(filter.Page, filter.PerPage, total)
	response.SuccessWithMeta(w, http.StatusOK, "Kasbon history retrieved", records, meta)
}

// RecordPayment records a kasbon payment
func (h *KasbonHandler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid customer ID")
		return
	}

	var input struct {
		Amount    int64   `json:"amount"`
		Notes     *string `json:"notes,omitempty"`
		CreatedBy *string `json:"created_by,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	v := validator.New()
	v.Positive("amount", input.Amount, "Amount must be positive")
	if v.HasErrors() {
		response.ValidationError(w, v.Errors())
		return
	}

	// Check if customer exists
	customer, err := h.customerRepo.GetByID(r.Context(), customerID)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Customer not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Failed to get customer")
		return
	}

	if customer.CurrentDebt == 0 {
		response.BadRequest(w, "Customer has no debt")
		return
	}

	paymentInput := domain.KasbonPaymentInput{
		CustomerID: customerID,
		Amount:     input.Amount,
		Notes:      input.Notes,
		CreatedBy:  input.CreatedBy,
	}

	record, err := h.kasbonRepo.CreatePayment(r.Context(), paymentInput)
	if err != nil {
		response.InternalServerError(w, "Failed to record payment")
		return
	}

	response.Created(w, "Payment recorded successfully", record)
}

// GetSummary retrieves kasbon summary for a customer
func (h *KasbonHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid customer ID")
		return
	}

	summary, err := h.kasbonRepo.GetSummary(r.Context(), customerID)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Customer not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Failed to get kasbon summary")
		return
	}

	response.OK(w, "Kasbon summary retrieved", summary)
}

// GetReport retrieves overall kasbon report
func (h *KasbonHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.kasbonRepo.GetReport(r.Context())
	if err != nil {
		response.InternalServerError(w, "Failed to get kasbon report")
		return
	}

	response.OK(w, "Kasbon report retrieved", report)
}

// DownloadBillingPDF generates and downloads billing PDF
func (h *KasbonHandler) DownloadBillingPDF(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid customer ID")
		return
	}

	// 1. Get Customer Info
	customer, err := h.customerRepo.GetByID(r.Context(), customerID)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Customer not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}

	// 2. Get Transactions (Last 30 days default, or query param)
	// Simplified: Get all active debts or last month history.
	// Let's grab last 30 days history to show detail.
	now := time.Now()
	dateFrom := now.AddDate(0, -1, 0) // 1 month ago
	
	filter := domain.KasbonFilter{
		DateFrom: &dateFrom,
		Page:     1,
		PerPage:  100, // reasonable limit for PDF
		// Sort by date asc for statement? Repository does DESC. We might want ASC.
		// Let's just take it and reverse in memory if needed or just show latest first.
		// Statements usually show Oldest -> Newest.
		// Repo sort DESC. So we should reverse it.
	}

	records, _, err := h.kasbonRepo.GetByCustomer(r.Context(), customerID, filter)
	if err != nil {
		response.InternalServerError(w, "Failed to get records")
		return
	}

	// Calculate Opening Balance (Current Debt - sum of mutations in period? or just 0 if all included?)
	// Accurate way: Get balance at DateFrom. Repository doesn't support that directly easily without `GetBalanceAt`.
	// Approximation: Opening Balance = CurrentDebt - (Debts in period - Payments in period).
	// Let's calculate backwards from current debt.
	
	currentDebt := customer.CurrentDebt
	
	// Reverse records to be chronological (Oldest first)
	// And calculate opening balance
	var history []domain.KasbonRecord
	for i := len(records) - 1; i >= 0; i-- {
		history = append(history, records[i])
	}
	
	// Calculate mutation in period
	var periodDebt, periodPay int64
	for _, rec := range history {
		if rec.Type == "debt" {
			periodDebt += rec.Amount
		} else {
			periodPay += rec.Amount
		}
	}
	
	openingBalance := currentDebt - (periodDebt - periodPay)

	// 3. Prepare PDF Data
	var billingTx []pdf.BillingTransaction
	runningBalance := openingBalance
	
	for _, rec := range history {
		desc := "Transaksi Kasbon"
		if rec.Notes != nil { desc = *rec.Notes }
		
		if rec.Type == "debt" {
			runningBalance += rec.Amount
		} else {
			runningBalance -= rec.Amount
		}
		
		billingTx = append(billingTx, pdf.BillingTransaction{
			Date:        rec.CreatedAt,
			Description: desc,
			Type:        string(rec.Type),
			Amount:      rec.Amount,
			Balance:     runningBalance,
		})
	}

	data := pdf.BillingData{
		StoreName:      "WARUNG KELONTONG", // Should get from Config
		StoreAddress:   "Jalan Raya No. 1",
		CustomerName:   customer.Name,
		InvoiceNumber:  fmt.Sprintf("INV/%s/%s", now.Format("200601"), customer.Name[:3]), // Pseudo
		Date:           now,
		PeriodStart:    &dateFrom,
		PeriodEnd:      &now,
		OpeningBalance: openingBalance,
		EndingBalance:  currentDebt,
		Transactions:   billingTx,
		PaymentInst:    "Silakan lakukan pembayaran ke Kasir atau Transfer BCA 1234567890 a.n Warung.",
	}

	// 4. Generate
	doc, err := pdf.GenerateBillingPDF(data)
	if err != nil {
		response.InternalServerError(w, "Failed to generate PDF")
		return
	}

	// 5. Stream Output
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=tagihan_%s.pdf", customer.Name))
	
	if err := doc.Output(w); err != nil {
		// Can't really write error response here if headers already sent
		// Log it
		fmt.Printf("Error writing PDF: %v\n", err)
	}
}
