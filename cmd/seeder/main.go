package main

import (
	"context"
	"log"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/pkg/password"
	"github.com/eveeze/warung-backend/internal/repository"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to PostgreSQL
	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Initialize repository
	userRepo := repository.NewUserRepository(db)
	ctx := context.Background()

	adminEmail := "admin@warung.com"
	defaultPassword := "password"
	hashedPassword, err := password.Hash(defaultPassword)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	admin := &domain.User{
		Name:         "Super Admin",
		Email:        adminEmail,
		PasswordHash: hashedPassword,
		Role:         domain.RoleAdmin,
		IsActive:     true,
	}

	// Check Admin
	if _, err := userRepo.GetByEmail(ctx, adminEmail); err != nil {
		// Create Admin
		if err := userRepo.Create(ctx, admin); err != nil {
			log.Printf("Failed to create admin user: %v\n", err)
		} else {
			log.Printf("Admin user created: %s / %s\n", adminEmail, defaultPassword)
		}
	} else {
		log.Println("Admin user already exists.")
	}

	// Create Cashier
	cashierEmail := "cashier@warung.com"
	if _, err := userRepo.GetByEmail(ctx, cashierEmail); err != nil {
		cashier := &domain.User{
			Name:         "Cashier One",
			Email:        cashierEmail,
			PasswordHash: hashedPassword,
			Role:         domain.RoleCashier,
			IsActive:     true,
		}
		if err := userRepo.Create(ctx, cashier); err != nil {
			log.Printf("Failed to create cashier: %v\n", err)
		} else {
			log.Printf("Cashier created: %s / %s\n", cashierEmail, defaultPassword)
		}
	} else {
		log.Println("Cashier user already exists.")
	}

	// Create Inventory
	inventoryEmail := "inventory@warung.com"
	if _, err := userRepo.GetByEmail(ctx, inventoryEmail); err != nil {
		inventory := &domain.User{
			Name:         "Inventory One",
			Email:        inventoryEmail,
			PasswordHash: hashedPassword,
			Role:         domain.RoleInventory,
			IsActive:     true,
		}
		if err := userRepo.Create(ctx, inventory); err != nil {
			log.Printf("Failed to create inventory user: %v\n", err)
		} else {
			log.Printf("Inventory User created: %s / %s\n", inventoryEmail, defaultPassword)
		}
	} else {
		log.Println("Inventory user already exists.")
	}

	// Create Products & Refillables (for testing)
	productRepo := repository.NewProductRepository(db)
	// refillRepo := repository.NewRefillableRepository(db) // Create method missing, use DB exec

	// 1. Empty Gas
	gasEmptyBarcode := "EMPTY3KG"
	gasEmptySKU := "GAS-3KG-EMPTY"
	gasEmptyStock := 10
	gasEmptyIsStock := true
	// gasEmptyActive := true (Unused in Input)
	gasEmptyInput := domain.ProductCreateInput{
		Name:          "Tabung Gas 3kg Kosong",
		Barcode:       &gasEmptyBarcode,
		SKU:           &gasEmptySKU,
		CurrentStock:  &gasEmptyStock,
		IsStockActive: &gasEmptyIsStock,
		Unit:          "tabung",
		BasePrice:     150000,
		CostPrice:     130000,
	}
	// Create returns *Product, error
	gasEmpty, err := productRepo.Create(ctx, gasEmptyInput)
	if err == nil {
		log.Printf("Created Empty Gas Product: %s\n", gasEmpty.ID)
	} else {
		log.Printf("Failed to create Empty Gas: %v\n", err)
	}

	// 2. Full Gas
	if gasEmpty != nil {
		gasFullBarcode := "GAS3KG"
		gasFullSKU := "GAS-3KG-FULL"
		gasFullStock := 50
		gasFullIsStock := true
		gasFullIsRefill := true
		// gasFullActive := true (Unused)
		
		gasFullInput := domain.ProductCreateInput{
			Name:           "Gas 3kg Isi",
			Barcode:        &gasFullBarcode,
			SKU:            &gasFullSKU,
			CurrentStock:   &gasFullStock,
			IsStockActive:  &gasFullIsStock,
			Unit:           "tabung",
			BasePrice:      20000,
			CostPrice:      18000,
			IsRefillable:   &gasFullIsRefill,
			EmptyProductID: &gasEmpty.ID,
		}

		gasFull, err := productRepo.Create(ctx, gasFullInput)
		if err == nil {
			log.Printf("Created Full Gas Product: %s\n", gasFull.ID)

			// Create Container Record (Direct SQL)
			// Table: refillable_containers (product_id, container_type, empty_count, full_count, notes)
			query := `
				INSERT INTO refillable_containers (product_id, container_type, empty_count, full_count, notes)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT DO NOTHING
			`
			_, err = db.ExecContext(ctx, query, gasFull.ID, "Gas 3kg", 10, 40, "Initial Seed")
			if err != nil {
				log.Printf("Failed to seed container: %v\n", err)
			} else {
				log.Printf("Seeded Refillable Container for %s\n", gasFull.Name)
			}
		} else {
			log.Printf("Failed to create Full Gas: %v\n", err)
		}
	}

	log.Printf("Seeding completed successfully.")
}
