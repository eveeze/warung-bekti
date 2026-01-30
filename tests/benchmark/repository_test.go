package benchmark

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/repository"
)

// Test configuration - uses environment variables or defaults
func getTestDBConfig() *config.DatabaseConfig {
	return &config.DatabaseConfig{
		Host:            getEnv("TEST_DB_HOST", "localhost"),
		Port:            getEnv("TEST_DB_PORT", "5432"),
		User:            getEnv("TEST_DB_USER", "postgres"),
		Password:        getEnv("TEST_DB_PASSWORD", "postgres"),
		DBName:          getEnv("TEST_DB_NAME", "warung_db"),
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// setupTestDB creates a test database connection
func setupTestDB(t testing.TB) *database.PostgresDB {
	cfg := getTestDBConfig()
	
	db, err := database.NewPostgres(cfg)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
	}

	return db
}

// ==================== PRODUCT TESTS ====================

// TestProductRepository_Create tests product creation
func TestProductRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewProductRepository(db)
	ctx := context.Background()

	// Test data
	input := domain.ProductCreateInput{
		Name:      "Test Product " + uuid.New().String()[:8],
		Unit:      "pcs",
		BasePrice: 10000,
		CostPrice: 7500,
	}

	// Test creation
	product, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// Assertions
	if product.ID == uuid.Nil {
		t.Error("Expected product ID to be set")
	}
	if product.Name != input.Name {
		t.Errorf("Expected name %s, got %s", input.Name, product.Name)
	}
	if product.BasePrice != input.BasePrice {
		t.Errorf("Expected price %d, got %d", input.BasePrice, product.BasePrice)
	}
	if !product.IsActive {
		t.Error("Expected product to be active")
	}

	// Cleanup
	_ = repo.Delete(ctx, product.ID)
}

// TestProductRepository_List tests product listing with filters
func TestProductRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewProductRepository(db)
	ctx := context.Background()

	// Test default list
	filter := domain.ProductFilter{Page: 1, PerPage: 20}
	products, total, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	t.Logf("Found %d products out of %d total", len(products), total)
}

// TestProductRepository_GetPricingTiersBatch tests batch loading
func TestProductRepository_GetPricingTiersBatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewProductRepository(db)
	ctx := context.Background()

	// Get some product IDs
	products, _, err := repo.List(ctx, domain.ProductFilter{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("Failed to list products: %v", err)
	}

	if len(products) == 0 {
		t.Skip("No products in database for testing")
	}

	productIDs := make([]uuid.UUID, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}

	// Test batch loading
	tiersMap, err := repo.GetPricingTiersBatch(ctx, productIDs)
	if err != nil {
		t.Fatalf("GetPricingTiersBatch failed: %v", err)
	}

	t.Logf("Loaded tiers for %d products", len(tiersMap))
}

// ==================== BENCHMARKS ====================

// BenchmarkProductRepository_List benchmarks product listing
func BenchmarkProductRepository_List(b *testing.B) {
	db := setupTestDB(b)
	defer db.Close()

	repo := repository.NewProductRepository(db)
	ctx := context.Background()

	filter := domain.ProductFilter{Page: 1, PerPage: 50}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := repo.List(ctx, filter)
		if err != nil {
			b.Fatalf("List failed: %v", err)
		}
	}
}

// BenchmarkProductRepository_Search benchmarks product search
func BenchmarkProductRepository_Search(b *testing.B) {
	db := setupTestDB(b)
	defer db.Close()

	repo := repository.NewProductRepository(db)
	ctx := context.Background()

	search := "mie"
	filter := domain.ProductFilter{
		Page:    1,
		PerPage: 20,
		Search:  &search,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := repo.List(ctx, filter)
		if err != nil {
			b.Fatalf("Search failed: %v", err)
		}
	}
}

// BenchmarkProductRepository_GetPricingTiersBatch benchmarks batch tier loading
func BenchmarkProductRepository_GetPricingTiersBatch(b *testing.B) {
	db := setupTestDB(b)
	defer db.Close()

	repo := repository.NewProductRepository(db)
	ctx := context.Background()

	// Get some product IDs
	products, _, err := repo.List(ctx, domain.ProductFilter{Page: 1, PerPage: 50})
	if err != nil {
		b.Skipf("Cannot get products: %v", err)
	}
	if len(products) == 0 {
		b.Skip("No products in database")
	}

	ids := make([]uuid.UUID, len(products))
	for i, p := range products {
		ids[i] = p.ID
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetPricingTiersBatch(ctx, ids)
		if err != nil {
			b.Fatalf("GetPricingTiersBatch failed: %v", err)
		}
	}
}

// BenchmarkProductRepository_GetByID benchmarks single product retrieval
func BenchmarkProductRepository_GetByID(b *testing.B) {
	db := setupTestDB(b)
	defer db.Close()

	repo := repository.NewProductRepository(db)
	ctx := context.Background()

	// Get a product ID to test with
	products, _, err := repo.List(ctx, domain.ProductFilter{Page: 1, PerPage: 1})
	if err != nil || len(products) == 0 {
		b.Skip("No products available for benchmark")
	}
	productID := products[0].ID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetByID(ctx, productID)
		if err != nil {
			b.Fatalf("GetByID failed: %v", err)
		}
	}
}

var _ = fmt.Sprintf // silence unused import
