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

	// Admin Only Helper
	adminOnly := func(h http.HandlerFunc) http.HandlerFunc {
		return authMiddleware(middleware.RequireAdmin()(http.HandlerFunc(h))).ServeHTTP
	}

	// Cashier Access Helper (Admin + Cashier)
	cashierAccess := func(h http.HandlerFunc) http.HandlerFunc {
		return authMiddleware(middleware.RequireRole("admin", "cashier")(http.HandlerFunc(h))).ServeHTTP
	}

	// Inventory Access Helper (Admin + Inventory)
	inventoryAccess := func(h http.HandlerFunc) http.HandlerFunc {
		return authMiddleware(middleware.RequireRole("admin", "inventory")(http.HandlerFunc(h))).ServeHTTP
	}

	// Products
	mux.HandleFunc("GET "+apiPrefix+"/products", protected(productHandler.List))
	mux.HandleFunc("POST "+apiPrefix+"/products", adminOnly(productHandler.Create)) // Admin only creation
	mux.HandleFunc("GET "+apiPrefix+"/products/search", protected(productHandler.GetByBarcode))
	mux.HandleFunc("GET "+apiPrefix+"/products/low-stock", protected(productHandler.GetLowStock))
	mux.HandleFunc("GET "+apiPrefix+"/products/{id}", protected(productHandler.GetByID))
	mux.HandleFunc("PUT "+apiPrefix+"/products/{id}", adminOnly(productHandler.Update)) // Admin only for modification
	mux.HandleFunc("DELETE "+apiPrefix+"/products/{id}", adminOnly(productHandler.Delete)) // Admin only
	mux.HandleFunc("POST "+apiPrefix+"/products/{id}/pricing-tiers", adminOnly(productHandler.AddPricingTier)) // Admin only
	mux.HandleFunc("PUT "+apiPrefix+"/products/{id}/pricing-tiers/{tierId}", adminOnly(productHandler.UpdatePricingTier))
	mux.HandleFunc("DELETE "+apiPrefix+"/products/{id}/pricing-tiers/{tierId}", adminOnly(productHandler.DeletePricingTier))

	// Customers
	mux.HandleFunc("GET "+apiPrefix+"/customers", cashierAccess(customerHandler.List))
	mux.HandleFunc("POST "+apiPrefix+"/customers", cashierAccess(customerHandler.Create))
	mux.HandleFunc("GET "+apiPrefix+"/customers/with-debt", cashierAccess(customerHandler.GetWithDebt))
	mux.HandleFunc("GET "+apiPrefix+"/customers/{id}", cashierAccess(customerHandler.GetByID))
	mux.HandleFunc("PUT "+apiPrefix+"/customers/{id}", cashierAccess(customerHandler.Update)) // Cashier may update customer info
	mux.HandleFunc("DELETE "+apiPrefix+"/customers/{id}", adminOnly(customerHandler.Delete)) // Admin only
	mux.HandleFunc("GET "+apiPrefix+"/customers/{id}/kasbon", cashierAccess(kasbonHandler.GetHistory))
	mux.HandleFunc("GET "+apiPrefix+"/customers/{id}/kasbon/summary", cashierAccess(kasbonHandler.GetSummary))
	mux.HandleFunc("POST "+apiPrefix+"/customers/{id}/kasbon/pay", cashierAccess(kasbonHandler.RecordPayment))

	// Transactions
	mux.HandleFunc("GET "+apiPrefix+"/transactions", cashierAccess(transactionHandler.List))
	mux.HandleFunc("POST "+apiPrefix+"/transactions", cashierAccess(transactionHandler.Create))
	mux.HandleFunc("POST "+apiPrefix+"/transactions/calculate", cashierAccess(transactionHandler.Calculate))
	mux.HandleFunc("GET "+apiPrefix+"/transactions/{id}", cashierAccess(transactionHandler.GetByID))
	mux.HandleFunc("POST "+apiPrefix+"/transactions/{id}/cancel", cashierAccess(transactionHandler.Cancel))

	// Inventory
	mux.HandleFunc("POST "+apiPrefix+"/inventory/restock", inventoryAccess(inventoryHandler.Restock)) // Inventory role allowed
	mux.HandleFunc("POST "+apiPrefix+"/inventory/adjust", adminOnly(inventoryHandler.Adjust)) // Admin only manual adjustment
	mux.HandleFunc("GET "+apiPrefix+"/inventory/low-stock", inventoryAccess(inventoryHandler.GetLowStock))
	mux.HandleFunc("GET "+apiPrefix+"/inventory/report", inventoryAccess(inventoryHandler.GetReport))
	mux.HandleFunc("GET "+apiPrefix+"/inventory/{productId}/movements", inventoryAccess(inventoryHandler.GetMovements))

	// Reports - Admin Only
	mux.HandleFunc("GET "+apiPrefix+"/reports/daily", adminOnly(reportHandler.GetDailyReport))
	mux.HandleFunc("GET "+apiPrefix+"/reports/kasbon", adminOnly(reportHandler.GetKasbonReport))
	mux.HandleFunc("GET "+apiPrefix+"/reports/inventory", adminOnly(reportHandler.GetInventoryReport))
	mux.HandleFunc("GET "+apiPrefix+"/reports/dashboard", adminOnly(reportHandler.GetDashboard))

	// Apply middleware chain
	var handler http.Handler = mux
	handler = middleware.Logging(handler)
	handler = middleware.CORS(handler)
	handler = middleware.RateLimit(100, time.Minute)(handler)
	handler = middleware.Recovery(handler)

	return handler
}
