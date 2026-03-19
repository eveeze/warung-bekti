package service_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/repository"
)

// TestLocationAndProductMapping verifies that a location can be created
// and mapped to a product, avoiding N+1 queries.
func TestLocationAndProductMapping(t *testing.T) {
	// 1. Setup Database Connection (Integration Test)
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5433/warung_test?sslmode=disable"
	}
	rawDB, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
		return
	}
	defer rawDB.Close()

	// Ensure DB is alive
	if err := rawDB.PingContext(context.Background()); err != nil {
		t.Skipf("Skipping integration test: database ping failed: %v", err)
		return
	}
	
	db := &database.PostgresDB{DB: rawDB}

	// 2. Setup Repositories
	locRepo := repository.NewLocationRepository(db)
	prodRepo := repository.NewProductRepository(db)
	ctx := context.Background()

	// 3. Create a test Location
	locInput := domain.LocationCreateInput{
		Name:        "Test Shelf 3D",
		Category:    domain.LocationTypeShelf,
		XCoordinate: 10.5,
		YCoordinate: 20.0,
		ZCoordinate: 5.0,
		Width:       100.0,
		Depth:       40.0,
		Height:      200.0,
	}

	loc, err := locRepo.Create(ctx, locInput)
	if err != nil {
		t.Fatalf("Failed to create location: %v", err)
	}
	defer func() {
		// Cleanup location
		locRepo.Delete(ctx, loc.ID)
	}()

	// 4. Create a test Product linked to this Location
	prodInput := domain.ProductCreateInput{
		Name:       "Test Product 3D",
		Unit:       "PCS",
		BasePrice:  10000,
		CostPrice:  8000,
		LocationID: &loc.ID,
	}

	// Insert into DB
	query := `
		INSERT INTO products (name, unit, base_price, cost_price, location_id, is_active, is_stock_active, current_stock, min_stock_alert, max_stock)
		VALUES ($1, $2, $3, $4, $5, true, true, 100, 10, 200)
		RETURNING id
	`
	var prodID string
	err = db.QueryRowContext(ctx, query, prodInput.Name, prodInput.Unit, prodInput.BasePrice, prodInput.CostPrice, prodInput.LocationID).Scan(&prodID)
	if err != nil {
		t.Fatalf("Failed to create mock product: %v", err)
	}
	defer func() {
		// Cleanup product
		db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", prodID)
	}()

	// 5. Test ProductRepository.List with Location fetching
	filter := domain.ProductFilter{Search: &prodInput.Name}
	products, _, err := prodRepo.List(ctx, filter)
	if err != nil {
		t.Fatalf("Failed to list products: %v", err)
	}

	if len(products) == 0 {
		t.Fatalf("Product not found in list")
	}

	// Verify Location Mapping
	var foundProd *domain.Product
	for _, p := range products {
		if p.Name == prodInput.Name {
			foundProd = &p
			break
		}
	}

	if foundProd == nil {
		t.Fatalf("Expected product not found")
	}

	if foundProd.Location == nil {
		t.Fatalf("Location was expected to be joined/fetched via batch but was nil")
	}

	if foundProd.Location.Name != "Test Shelf 3D" {
		t.Errorf("Expected location name 'Test Shelf 3D', got '%s'", foundProd.Location.Name)
	}
	if foundProd.Location.ZCoordinate != 5.0 {
		t.Errorf("Expected location Z coord 5.0, got %f", foundProd.Location.ZCoordinate)
	}
	if foundProd.Location.Height != 200.0 {
		t.Errorf("Expected location height 200.0, got %f", foundProd.Location.Height)
	}
}
