package router

import (
	"net/http"
	"time"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/handler"
	"github.com/eveeze/warung-backend/internal/middleware"
	"github.com/eveeze/warung-backend/internal/repository"
	"github.com/eveeze/warung-backend/internal/service"
	"github.com/eveeze/warung-backend/internal/storage"
)

// New creates and configures the HTTP router
func New(
	cfg *config.Config,
	db *database.PostgresDB,
	redis *database.RedisClient,
	minio *storage.MinioClient,
) http.Handler {
	mux := http.NewServeMux()

	// Initialize repositories
	productRepo := repository.NewProductRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	kasbonRepo := repository.NewKasbonRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)

	// Initialize services
	transactionSvc := service.NewTransactionService(
		db, transactionRepo, productRepo, customerRepo, kasbonRepo, inventoryRepo,
	)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(db, redis, minio)
	productHandler := handler.NewProductHandler(productRepo)
	customerHandler := handler.NewCustomerHandler(customerRepo)
	transactionHandler := handler.NewTransactionHandler(transactionSvc, transactionRepo)
	kasbonHandler := handler.NewKasbonHandler(kasbonRepo, customerRepo)
	inventoryHandler := handler.NewInventoryHandler(inventoryRepo, productRepo)
	reportHandler := handler.NewReportHandler(transactionRepo, kasbonRepo, inventoryRepo, productRepo)

	// Health check routes
	mux.HandleFunc("GET /health", healthHandler.Health)
	mux.HandleFunc("GET /ready", healthHandler.Ready)
	mux.HandleFunc("GET /live", healthHandler.Live)

	// API v1 routes
	apiPrefix := "/api/v1"

	// Products
	mux.HandleFunc("GET "+apiPrefix+"/products", productHandler.List)
	mux.HandleFunc("POST "+apiPrefix+"/products", productHandler.Create)
	mux.HandleFunc("GET "+apiPrefix+"/products/search", productHandler.GetByBarcode)
	mux.HandleFunc("GET "+apiPrefix+"/products/low-stock", productHandler.GetLowStock)
	mux.HandleFunc("GET "+apiPrefix+"/products/{id}", productHandler.GetByID)
	mux.HandleFunc("PUT "+apiPrefix+"/products/{id}", productHandler.Update)
	mux.HandleFunc("DELETE "+apiPrefix+"/products/{id}", productHandler.Delete)
	mux.HandleFunc("POST "+apiPrefix+"/products/{id}/pricing-tiers", productHandler.AddPricingTier)
	mux.HandleFunc("PUT "+apiPrefix+"/products/{id}/pricing-tiers/{tierId}", productHandler.UpdatePricingTier)
	mux.HandleFunc("DELETE "+apiPrefix+"/products/{id}/pricing-tiers/{tierId}", productHandler.DeletePricingTier)

	// Customers
	mux.HandleFunc("GET "+apiPrefix+"/customers", customerHandler.List)
	mux.HandleFunc("POST "+apiPrefix+"/customers", customerHandler.Create)
	mux.HandleFunc("GET "+apiPrefix+"/customers/with-debt", customerHandler.GetWithDebt)
	mux.HandleFunc("GET "+apiPrefix+"/customers/{id}", customerHandler.GetByID)
	mux.HandleFunc("PUT "+apiPrefix+"/customers/{id}", customerHandler.Update)
	mux.HandleFunc("DELETE "+apiPrefix+"/customers/{id}", customerHandler.Delete)
	mux.HandleFunc("GET "+apiPrefix+"/customers/{id}/kasbon", kasbonHandler.GetHistory)
	mux.HandleFunc("GET "+apiPrefix+"/customers/{id}/kasbon/summary", kasbonHandler.GetSummary)
	mux.HandleFunc("POST "+apiPrefix+"/customers/{id}/kasbon/pay", kasbonHandler.RecordPayment)

	// Transactions
	mux.HandleFunc("GET "+apiPrefix+"/transactions", transactionHandler.List)
	mux.HandleFunc("POST "+apiPrefix+"/transactions", transactionHandler.Create)
	mux.HandleFunc("POST "+apiPrefix+"/transactions/calculate", transactionHandler.Calculate)
	mux.HandleFunc("GET "+apiPrefix+"/transactions/{id}", transactionHandler.GetByID)
	mux.HandleFunc("POST "+apiPrefix+"/transactions/{id}/cancel", transactionHandler.Cancel)

	// Inventory
	mux.HandleFunc("POST "+apiPrefix+"/inventory/restock", inventoryHandler.Restock)
	mux.HandleFunc("POST "+apiPrefix+"/inventory/adjust", inventoryHandler.Adjust)
	mux.HandleFunc("GET "+apiPrefix+"/inventory/low-stock", inventoryHandler.GetLowStock)
	mux.HandleFunc("GET "+apiPrefix+"/inventory/report", inventoryHandler.GetReport)
	mux.HandleFunc("GET "+apiPrefix+"/inventory/{productId}/movements", inventoryHandler.GetMovements)

	// Reports
	mux.HandleFunc("GET "+apiPrefix+"/reports/daily", reportHandler.GetDailyReport)
	mux.HandleFunc("GET "+apiPrefix+"/reports/kasbon", reportHandler.GetKasbonReport)
	mux.HandleFunc("GET "+apiPrefix+"/reports/inventory", reportHandler.GetInventoryReport)
	mux.HandleFunc("GET "+apiPrefix+"/reports/dashboard", reportHandler.GetDashboard)

	// Apply middleware chain
	var handler http.Handler = mux
	handler = middleware.Logging(handler)
	handler = middleware.CORS(handler)
	handler = middleware.RateLimit(100, time.Minute)(handler)
	handler = middleware.Recovery(handler)

	return handler
}
