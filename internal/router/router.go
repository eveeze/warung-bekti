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
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	transactionSvc := service.NewTransactionService(
		db, transactionRepo, productRepo, customerRepo, kasbonRepo, inventoryRepo,
	)
	authSvc := service.NewAuthService(userRepo, cfg)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(db, redis, minio)
	productHandler := handler.NewProductHandler(productRepo)
	customerHandler := handler.NewCustomerHandler(customerRepo)
	transactionHandler := handler.NewTransactionHandler(transactionSvc, transactionRepo)
	kasbonHandler := handler.NewKasbonHandler(kasbonRepo, customerRepo)
	inventoryHandler := handler.NewInventoryHandler(inventoryRepo, productRepo)
	reportHandler := handler.NewReportHandler(transactionRepo, kasbonRepo, inventoryRepo, productRepo)
	authHandler := handler.NewAuthHandler(authSvc)

	// Health check routes (Public)
	mux.HandleFunc("GET /health", healthHandler.Health)
	mux.HandleFunc("GET /ready", healthHandler.Ready)
	mux.HandleFunc("GET /live", healthHandler.Live)

	// Auth routes (Public)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/refresh", authHandler.RefreshToken)

	// API v1 routes
	apiPrefix := "/api/v1"

	// Middleware for protected routes
	authMiddleware := middleware.Auth(&cfg.JWT)

	// Protected Routes Helper
	// We wrap handlers with authMiddleware
	protected := func(h http.HandlerFunc) http.HandlerFunc {
		return authMiddleware(http.HandlerFunc(h)).ServeHTTP
	}

	// Products
	mux.HandleFunc("GET "+apiPrefix+"/products", protected(productHandler.List))
	mux.HandleFunc("POST "+apiPrefix+"/products", protected(productHandler.Create))
	mux.HandleFunc("GET "+apiPrefix+"/products/search", protected(productHandler.GetByBarcode))
	mux.HandleFunc("GET "+apiPrefix+"/products/low-stock", protected(productHandler.GetLowStock))
	mux.HandleFunc("GET "+apiPrefix+"/products/{id}", protected(productHandler.GetByID))
	mux.HandleFunc("PUT "+apiPrefix+"/products/{id}", protected(productHandler.Update))
	mux.HandleFunc("DELETE "+apiPrefix+"/products/{id}", protected(productHandler.Delete))
	mux.HandleFunc("POST "+apiPrefix+"/products/{id}/pricing-tiers", protected(productHandler.AddPricingTier))
	mux.HandleFunc("PUT "+apiPrefix+"/products/{id}/pricing-tiers/{tierId}", protected(productHandler.UpdatePricingTier))
	mux.HandleFunc("DELETE "+apiPrefix+"/products/{id}/pricing-tiers/{tierId}", protected(productHandler.DeletePricingTier))

	// Customers
	mux.HandleFunc("GET "+apiPrefix+"/customers", protected(customerHandler.List))
	mux.HandleFunc("POST "+apiPrefix+"/customers", protected(customerHandler.Create))
	mux.HandleFunc("GET "+apiPrefix+"/customers/with-debt", protected(customerHandler.GetWithDebt))
	mux.HandleFunc("GET "+apiPrefix+"/customers/{id}", protected(customerHandler.GetByID))
	mux.HandleFunc("PUT "+apiPrefix+"/customers/{id}", protected(customerHandler.Update))
	mux.HandleFunc("DELETE "+apiPrefix+"/customers/{id}", protected(customerHandler.Delete))
	mux.HandleFunc("GET "+apiPrefix+"/customers/{id}/kasbon", protected(kasbonHandler.GetHistory))
	mux.HandleFunc("GET "+apiPrefix+"/customers/{id}/kasbon/summary", protected(kasbonHandler.GetSummary))
	mux.HandleFunc("POST "+apiPrefix+"/customers/{id}/kasbon/pay", protected(kasbonHandler.RecordPayment))

	// Transactions
	mux.HandleFunc("GET "+apiPrefix+"/transactions", protected(transactionHandler.List))
	mux.HandleFunc("POST "+apiPrefix+"/transactions", protected(transactionHandler.Create))
	mux.HandleFunc("POST "+apiPrefix+"/transactions/calculate", protected(transactionHandler.Calculate))
	mux.HandleFunc("GET "+apiPrefix+"/transactions/{id}", protected(transactionHandler.GetByID))
	mux.HandleFunc("POST "+apiPrefix+"/transactions/{id}/cancel", protected(transactionHandler.Cancel))

	// Inventory
	mux.HandleFunc("POST "+apiPrefix+"/inventory/restock", protected(inventoryHandler.Restock))
	mux.HandleFunc("POST "+apiPrefix+"/inventory/adjust", protected(inventoryHandler.Adjust))
	mux.HandleFunc("GET "+apiPrefix+"/inventory/low-stock", protected(inventoryHandler.GetLowStock))
	mux.HandleFunc("GET "+apiPrefix+"/inventory/report", protected(inventoryHandler.GetReport))
	mux.HandleFunc("GET "+apiPrefix+"/inventory/{productId}/movements", protected(inventoryHandler.GetMovements))

	// Reports
	mux.HandleFunc("GET "+apiPrefix+"/reports/daily", protected(reportHandler.GetDailyReport))
	mux.HandleFunc("GET "+apiPrefix+"/reports/kasbon", protected(reportHandler.GetKasbonReport))
	mux.HandleFunc("GET "+apiPrefix+"/reports/inventory", protected(reportHandler.GetInventoryReport))
	mux.HandleFunc("GET "+apiPrefix+"/reports/dashboard", protected(reportHandler.GetDashboard))

	// Apply middleware chain
	var handler http.Handler = mux
	handler = middleware.Logging(handler)
	handler = middleware.CORS(handler)
	handler = middleware.RateLimit(100, time.Minute)(handler)
	handler = middleware.Recovery(handler)

	return handler
}
