package handler

import (
	"net/http"
	"time"

	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/repository"
)

// ReportHandler handles report endpoints
type ReportHandler struct {
	transactionRepo *repository.TransactionRepository
	kasbonRepo      *repository.KasbonRepository
	inventoryRepo   *repository.InventoryRepository
	productRepo     *repository.ProductRepository
}

// NewReportHandler creates a new ReportHandler
func NewReportHandler(
	transactionRepo *repository.TransactionRepository,
	kasbonRepo *repository.KasbonRepository,
	inventoryRepo *repository.InventoryRepository,
	productRepo *repository.ProductRepository,
) *ReportHandler {
	return &ReportHandler{
		transactionRepo: transactionRepo,
		kasbonRepo:      kasbonRepo,
		inventoryRepo:   inventoryRepo,
		productRepo:     productRepo,
	}
}

// DailyReportSummary represents summary metrics
type DailyReportSummary struct {
	Date               string `json:"date"`
	TotalSales         int64  `json:"total_sales"`
	TotalTransactions  int    `json:"total_transactions"`
	EstimatedProfit    int64  `json:"estimated_profit"`
	AverageTransaction int64  `json:"average_transaction"`
	TotalProfit        int64  `json:"total_profit"`
}

// DailyReportResponse represents the full daily report response
type DailyReportResponse struct {
	Summary      DailyReportSummary       `json:"summary"`
	HourlySales  []map[string]interface{} `json:"hourly_sales"`
	TopProducts  []map[string]interface{} `json:"top_products"`
}

// GetDailyReport returns daily sales summary with details
func (h *ReportHandler) GetDailyReport(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	sales, count, err := h.transactionRepo.GetDailySales(r.Context(), dateStr)
	if err != nil {
		response.InternalServerError(w, "Failed to get daily sales")
		return
	}

	profit, err := h.transactionRepo.GetDailyProfit(r.Context(), dateStr)
	if err != nil {
		profit = 0
	}

	hourly, err := h.transactionRepo.GetHourlySales(r.Context(), dateStr)
	if err != nil {
		hourly = []map[string]interface{}{}
	}

	topProducts, err := h.transactionRepo.GetTopProducts(r.Context(), dateStr, 5)
	if err != nil {
		topProducts = []map[string]interface{}{}
	}

	avgTransaction := int64(0)
	if count > 0 {
		avgTransaction = sales / int64(count)
	}

	report := DailyReportResponse{
		Summary: DailyReportSummary{
			Date:               dateStr,
			TotalSales:         sales,
			TotalTransactions:  count,
			EstimatedProfit:    profit,
			AverageTransaction: avgTransaction,
			TotalProfit:        profit,
		},
		HourlySales: hourly,
		TopProducts: topProducts,
	}

	response.OK(w, "Daily report retrieved", report)
}

// GetKasbonReport returns kasbon summary
func (h *ReportHandler) GetKasbonReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.kasbonRepo.GetReport(r.Context())
	if err != nil {
		response.InternalServerError(w, "Failed to get kasbon report")
		return
	}

	response.OK(w, "Kasbon report retrieved", report)
}

// GetInventoryReport returns stock inventory summary
func (h *ReportHandler) GetInventoryReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.inventoryRepo.GetStockReport(r.Context())
	if err != nil {
		response.InternalServerError(w, "Failed to get inventory report")
		return
	}

	lowStock, _ := h.productRepo.GetLowStockProducts(r.Context())
	report.LowStockProducts = lowStock

	response.OK(w, "Inventory report retrieved", report)
}

// Dashboard represents dashboard summary
type Dashboard struct {
	Today           DailyReportSummary `json:"today"`
	TotalOutstanding int64             `json:"total_outstanding_kasbon"`
	LowStockCount   int                `json:"low_stock_count"`
	OutOfStockCount int                `json:"out_of_stock_count"`
}

// GetDashboard returns dashboard summary
func (h *ReportHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	today := time.Now().Format("2006-01-02")

	sales, count, _ := h.transactionRepo.GetDailySales(r.Context(), today)
	profit, _ := h.transactionRepo.GetDailyProfit(r.Context(), today)

	kasbonReport, _ := h.kasbonRepo.GetReport(r.Context())
	stockReport, _ := h.inventoryRepo.GetStockReport(r.Context())

	avgTransaction := int64(0)
	if count > 0 {
		avgTransaction = sales / int64(count)
	}

	dashboard := Dashboard{
		Today: DailyReportSummary{
			Date:               today,
			TotalSales:         sales,
			TotalTransactions:  count,
			EstimatedProfit:    profit,
			AverageTransaction: avgTransaction,
			TotalProfit:        profit,
		},
	}

	if kasbonReport != nil {
		dashboard.TotalOutstanding = kasbonReport.TotalOutstanding
	}
	if stockReport != nil {
		dashboard.LowStockCount = stockReport.LowStockCount
		dashboard.OutOfStockCount = stockReport.OutOfStockCount
	}

	response.OK(w, "Dashboard retrieved", dashboard)
}
