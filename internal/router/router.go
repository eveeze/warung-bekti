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
	paymentRepo := repository.NewPaymentRepository(db)
	stockOpnameRepo := repository.NewStockOpnameRepository(db)
	cashFlowRepo := repository.NewCashFlowRepository(db)
	posRepo := repository.NewPOSRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	consignmentRepo := repository.NewConsignmentRepository(db)
	refillableRepo := repository.NewRefillableRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	// Initialize services
	transactionSvc := service.NewTransactionService(
		db, transactionRepo, productRepo, customerRepo, kasbonRepo, inventoryRepo,
	)
	authSvc := service.NewAuthService(userRepo, cfg)
	userSvc := service.NewUserService(userRepo) // New Service initialized
	paymentSvc := service.NewPaymentService(db, paymentRepo, transactionRepo, &cfg.Midtrans)
	stockOpnameSvc := service.NewStockOpnameService(db, stockOpnameRepo, productRepo, inventoryRepo)
	cashFlowSvc := service.NewCashFlowService(db, cashFlowRepo)
	posSvc := service.NewPOSService(db, posRepo, productRepo, transactionRepo, inventoryRepo)
	consignmentSvc := service.NewConsignmentService(db, consignmentRepo, transactionRepo)
	refillableSvc := service.NewRefillableService(db, refillableRepo)
	categorySvc := service.NewCategoryService(categoryRepo)

	// Initialize cache service
	cacheSvc := service.NewCacheService(redis)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(db, redis, minio)
	productHandler := handler.NewProductHandler(productRepo, minio, cacheSvc)
	customerHandler := handler.NewCustomerHandler(customerRepo)
	transactionHandler := handler.NewTransactionHandler(transactionSvc, transactionRepo)
	kasbonHandler := handler.NewKasbonHandler(kasbonRepo, customerRepo)
	inventoryHandler := handler.NewInventoryHandler(inventoryRepo, productRepo)
	reportHandler := handler.NewReportHandler(transactionRepo, kasbonRepo, inventoryRepo, productRepo)
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc) // New Handler initialized
	paymentHandler := handler.NewPaymentHandler(paymentSvc)
	stockOpnameHandler := handler.NewStockOpnameHandler(stockOpnameSvc)
	cashFlowHandler := handler.NewCashFlowHandler(cashFlowSvc)
	posHandler := handler.NewPOSHandler(posSvc)
	consignmentHandler := handler.NewConsignmentHandler(consignmentSvc)
	refillableHandler := handler.NewRefillableHandler(refillableSvc)
	categoryHandler := handler.NewCategoryHandler(categorySvc)

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

	// Helpers for Middleware wrapping
	protected := func(h http.HandlerFunc) http.HandlerFunc {
		return authMiddleware(http.HandlerFunc(h)).ServeHTTP
	}

	adminOnly := func(h http.HandlerFunc) http.HandlerFunc {
		return authMiddleware(middleware.RequireAdmin()(http.HandlerFunc(h))).ServeHTTP
	}

	cashierAccess := func(h http.HandlerFunc) http.HandlerFunc {
		return authMiddleware(middleware.RequireRole("admin", "cashier")(http.HandlerFunc(h))).ServeHTTP
	}

	inventoryAccess := func(h http.HandlerFunc) http.HandlerFunc {
		return authMiddleware(middleware.RequireRole("admin", "inventory")(http.HandlerFunc(h))).ServeHTTP
	}

	// ========================================================================
	// USERS MANAGEMENT (Admin Only)
	// ========================================================================
	// Frontend: client.get('/users')
	mux.HandleFunc("GET "+apiPrefix+"/users", adminOnly(userHandler.List))
	
	// Frontend: client.post('/users')
	mux.HandleFunc("POST "+apiPrefix+"/users", adminOnly(userHandler.Create))
	
	// Frontend: client.get('/users/:id'), client.put('/users/:id'), client.delete('/users/:id')
	mux.HandleFunc("GET "+apiPrefix+"/users/{id}", adminOnly(userHandler.GetByID))
	mux.HandleFunc("PUT "+apiPrefix+"/users/{id}", adminOnly(userHandler.Update))
	mux.HandleFunc("DELETE "+apiPrefix+"/users/{id}", adminOnly(userHandler.Delete))

	// ========================================================================
	// OTHER MODULES
	// ========================================================================

	// Products
	mux.HandleFunc("GET "+apiPrefix+"/products", protected(productHandler.List))
	mux.HandleFunc("POST "+apiPrefix+"/products", adminOnly(productHandler.Create))
	mux.HandleFunc("GET "+apiPrefix+"/products/search", protected(productHandler.GetByBarcode))
	mux.HandleFunc("GET "+apiPrefix+"/products/low-stock", protected(productHandler.GetLowStock))
	mux.HandleFunc("GET "+apiPrefix+"/products/{id}", protected(productHandler.GetByID))
	mux.HandleFunc("PUT "+apiPrefix+"/products/{id}", adminOnly(productHandler.Update))
	mux.HandleFunc("DELETE "+apiPrefix+"/products/{id}", adminOnly(productHandler.Delete))
	mux.HandleFunc("POST "+apiPrefix+"/products/{id}/pricing-tiers", adminOnly(productHandler.AddPricingTier))
	mux.HandleFunc("PUT "+apiPrefix+"/products/{id}/pricing-tiers/{tierId}", adminOnly(productHandler.UpdatePricingTier))
	mux.HandleFunc("DELETE "+apiPrefix+"/products/{id}/pricing-tiers/{tierId}", adminOnly(productHandler.DeletePricingTier))

	// Customers
	mux.HandleFunc("GET "+apiPrefix+"/customers", cashierAccess(customerHandler.List))
	mux.HandleFunc("POST "+apiPrefix+"/customers", cashierAccess(customerHandler.Create))
	mux.HandleFunc("GET "+apiPrefix+"/customers/with-debt", cashierAccess(customerHandler.GetWithDebt))
	mux.HandleFunc("GET "+apiPrefix+"/customers/{id}", cashierAccess(customerHandler.GetByID))
	mux.HandleFunc("PUT "+apiPrefix+"/customers/{id}", cashierAccess(customerHandler.Update))
	mux.HandleFunc("DELETE "+apiPrefix+"/customers/{id}", adminOnly(customerHandler.Delete))
	mux.HandleFunc("GET "+apiPrefix+"/kasbon/customers/{id}", cashierAccess(kasbonHandler.GetHistory))
	mux.HandleFunc("GET "+apiPrefix+"/kasbon/customers/{id}/summary", cashierAccess(kasbonHandler.GetSummary))
	mux.HandleFunc("GET "+apiPrefix+"/kasbon/customers/{id}/billing/pdf", cashierAccess(kasbonHandler.DownloadBillingPDF))
	mux.HandleFunc("POST "+apiPrefix+"/kasbon/customers/{id}/payments", cashierAccess(kasbonHandler.RecordPayment))

	// Transactions
	mux.HandleFunc("GET "+apiPrefix+"/transactions", cashierAccess(transactionHandler.List))
	mux.HandleFunc("POST "+apiPrefix+"/transactions", cashierAccess(transactionHandler.Create))
	mux.HandleFunc("POST "+apiPrefix+"/transactions/calculate", cashierAccess(transactionHandler.Calculate))
	mux.HandleFunc("GET "+apiPrefix+"/transactions/{id}", cashierAccess(transactionHandler.GetByID))
	mux.HandleFunc("POST "+apiPrefix+"/transactions/{id}/cancel", cashierAccess(transactionHandler.Cancel))

	// Inventory
	mux.HandleFunc("POST "+apiPrefix+"/inventory/restock", inventoryAccess(inventoryHandler.Restock))
	mux.HandleFunc("POST "+apiPrefix+"/inventory/adjust", adminOnly(inventoryHandler.Adjust))
	mux.HandleFunc("GET "+apiPrefix+"/inventory/low-stock", inventoryAccess(inventoryHandler.GetLowStock))
	mux.HandleFunc("GET "+apiPrefix+"/inventory/report", inventoryAccess(inventoryHandler.GetReport))
	mux.HandleFunc("GET "+apiPrefix+"/inventory/restock-list/pdf", inventoryAccess(inventoryHandler.DownloadRestockPDF))
	mux.HandleFunc("GET "+apiPrefix+"/inventory/{productId}/movements", inventoryAccess(inventoryHandler.GetMovements))

	// Categories
	mux.HandleFunc("GET "+apiPrefix+"/categories", protected(categoryHandler.List))
	mux.HandleFunc("GET "+apiPrefix+"/categories/{id}", protected(categoryHandler.GetByID))
	mux.HandleFunc("POST "+apiPrefix+"/categories", adminOnly(categoryHandler.Create))
	mux.HandleFunc("PUT "+apiPrefix+"/categories/{id}", adminOnly(categoryHandler.Update))
	mux.HandleFunc("DELETE "+apiPrefix+"/categories/{id}", adminOnly(categoryHandler.Delete))

	// Reports
	mux.HandleFunc("GET "+apiPrefix+"/reports/daily", adminOnly(reportHandler.GetDailyReport))
	mux.HandleFunc("GET "+apiPrefix+"/reports/kasbon", adminOnly(reportHandler.GetKasbonReport))
	mux.HandleFunc("GET "+apiPrefix+"/reports/inventory", adminOnly(reportHandler.GetInventoryReport))
	mux.HandleFunc("GET "+apiPrefix+"/reports/dashboard", adminOnly(reportHandler.GetDashboard))

	// Payments
	mux.HandleFunc("POST "+apiPrefix+"/payments/snap", cashierAccess(paymentHandler.GenerateSnapToken))
	mux.HandleFunc("POST "+apiPrefix+"/payments/notification", paymentHandler.HandleNotification)
	mux.HandleFunc("POST "+apiPrefix+"/payments/{id}/manual-verify", adminOnly(paymentHandler.ManualVerify))
	mux.HandleFunc("GET "+apiPrefix+"/payments/transaction/{id}", cashierAccess(paymentHandler.GetPaymentByTransaction))

	// Stock Opname
	mux.HandleFunc("POST "+apiPrefix+"/stock-opname/sessions", inventoryAccess(stockOpnameHandler.StartSession))
	mux.HandleFunc("GET "+apiPrefix+"/stock-opname/sessions", inventoryAccess(stockOpnameHandler.ListSessions))
	mux.HandleFunc("GET "+apiPrefix+"/stock-opname/sessions/{id}", inventoryAccess(stockOpnameHandler.GetSession))
	mux.HandleFunc("POST "+apiPrefix+"/stock-opname/sessions/{id}/items", inventoryAccess(stockOpnameHandler.RecordCount))
	mux.HandleFunc("POST "+apiPrefix+"/stock-opname/sessions/{id}/finalize", inventoryAccess(stockOpnameHandler.FinalizeSession))
	mux.HandleFunc("GET "+apiPrefix+"/stock-opname/sessions/{id}/variance", inventoryAccess(stockOpnameHandler.GetVarianceReport))
	mux.HandleFunc("POST "+apiPrefix+"/stock-opname/sessions/{id}/cancel", inventoryAccess(stockOpnameHandler.CancelSession))
	mux.HandleFunc("GET "+apiPrefix+"/stock-opname/shopping-list", inventoryAccess(stockOpnameHandler.GetShoppingList))
	mux.HandleFunc("GET "+apiPrefix+"/stock-opname/near-expiry", inventoryAccess(stockOpnameHandler.GetNearExpiryReport))

	// Cash Flow
	mux.HandleFunc("POST "+apiPrefix+"/cashflow/drawer/open", cashierAccess(cashFlowHandler.OpenDrawer))
	mux.HandleFunc("POST "+apiPrefix+"/cashflow/drawer/close", cashierAccess(cashFlowHandler.CloseDrawer))
	mux.HandleFunc("GET "+apiPrefix+"/cashflow/drawer/current", cashierAccess(cashFlowHandler.GetCurrentSession))
	mux.HandleFunc("GET "+apiPrefix+"/cashflow/categories", cashierAccess(cashFlowHandler.GetCategories))
	mux.HandleFunc("POST "+apiPrefix+"/cashflow", cashierAccess(cashFlowHandler.RecordCashFlow))
	mux.HandleFunc("GET "+apiPrefix+"/cashflow", cashierAccess(cashFlowHandler.ListCashFlows))

	// POS Features
	mux.HandleFunc("POST "+apiPrefix+"/pos/held-carts", cashierAccess(posHandler.HoldCart))
	mux.HandleFunc("GET "+apiPrefix+"/pos/held-carts", cashierAccess(posHandler.ListHeldCarts))
	mux.HandleFunc("GET "+apiPrefix+"/pos/held-carts/{id}", cashierAccess(posHandler.GetHeldCart))
	mux.HandleFunc("POST "+apiPrefix+"/pos/held-carts/{id}/resume", cashierAccess(posHandler.ResumeCart))
	mux.HandleFunc("POST "+apiPrefix+"/pos/held-carts/{id}/discard", cashierAccess(posHandler.DiscardCart))
	mux.HandleFunc("POST "+apiPrefix+"/pos/refunds", cashierAccess(posHandler.CreateRefund))

	// Consignment
	mux.HandleFunc("POST "+apiPrefix+"/consignors", adminOnly(consignmentHandler.CreateConsignor))
	mux.HandleFunc("GET "+apiPrefix+"/consignors", adminOnly(consignmentHandler.ListConsignors))
	mux.HandleFunc("PUT "+apiPrefix+"/consignors/{id}", adminOnly(consignmentHandler.UpdateConsignor))

	// Refillables
	mux.HandleFunc("GET "+apiPrefix+"/refillables", inventoryAccess(refillableHandler.GetContainers))
	mux.HandleFunc("POST "+apiPrefix+"/refillables/adjust", inventoryAccess(refillableHandler.AdjustStock))

	// Apply global middleware chain
	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.Audit(auditRepo)(h)
	h = middleware.CORS(h)
	h = middleware.RateLimit(1000, time.Minute)(h)
	h = middleware.Recovery(h)

	return h
}