package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/pkg/password"
	"github.com/eveeze/warung-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
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

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	ctx := context.Background()

	// --- Seed Users ---
	seedUsers(ctx, userRepo)

	// --- Seed Products from Excel ---
	excelFile := "ipos_data.xlsx"
	log.Printf("Starting seeding from %s...", excelFile)
	if err := seedProductsFromExcel(ctx, db, productRepo, categoryRepo, excelFile); err != nil {
		log.Printf("Error seeding from Excel: %v", err)
	} else {
		log.Println("Seeding from Excel completed successfully.")
	}
}

func seedUsers(ctx context.Context, userRepo domain.UserRepository) {
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
}

func seedProductsFromExcel(ctx context.Context, db *database.PostgresDB, productRepo *repository.ProductRepository, categoryRepo *repository.CategoryRepository, filename string) error {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	// Get the first sheet name
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return fmt.Errorf("no sheets found in excel file")
	}
	sheetName := sheets[0]
	log.Printf("Reading sheet: %s", sheetName)

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) < 2 {
		return fmt.Errorf("excel file is empty or missing header")
	}

	// Parse header to find column indices
	if len(rows) < 3 {
		return fmt.Errorf("excel file is empty or missing header")
	}
	header := rows[1]
	colMap := make(map[string]int)
	for i, col := range header {
		// Clean the header column name for robust matching:
		// 1. Convert to Uppercase
		// 2. Trim spaces
		key := strings.TrimSpace(strings.ToUpper(col))
		colMap[key] = i
	}

	// Helper to get value
	getValue := func(row []string, colName string) string {
		idx, ok := colMap[colName]
		if !ok || idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[idx])
	}

	// Helper to get string pointer or nil if empty
	getStringPtr := func(s string) *string {
		if s == "" || s == "NA" || s == "-" {
			return nil
		}
		return &s
	}

	// Cache categories to avoid DB hits
	categoryCache := make(map[string]uuid.UUID)

	// Preload categories
	cats, err := categoryRepo.FindAll(ctx, nil)
	if err == nil {
		for _, c := range cats {
			categoryCache[c.Name] = c.ID
		}
	}

	successCount := 0
	duplicateCount := 0
	errorCount := 0
	skippedCount := 0

	// Iterate rows (skip header)
	for i, row := range rows[2:] {
		// Basic fields
		sku := getValue(row, "KODE ITEM")
		barcode := getValue(row, "BARCODE")
		name := getValue(row, "NAMA ITEM")
		categoryName := getValue(row, "JENIS")
		unit := getValue(row, "SATUAN")
		
		if name == "" {
			skippedCount++
			continue 
		}

		// Check/Create Category
		var categoryID *uuid.UUID
		if categoryName != "" {
			if id, ok := categoryCache[categoryName]; ok {
				categoryID = &id
			} else {
				// Create new
				newCat := domain.Category{Name: categoryName}
				if err := categoryRepo.Create(ctx, &newCat); err == nil {
					categoryCache[categoryName] = newCat.ID
					categoryID = &newCat.ID
				} else {
					// Proceed without category id
				}
			}
		}

		// Prices
		costPriceStr := cleanNumber(getValue(row, "HARGA POKOK"))
		sellPriceStr := cleanNumber(getValue(row, "HARGA JUAL"))
		currentStockStr := cleanNumber(getValue(row, "STOK AWAL"))
		minStockStr := cleanNumber(getValue(row, "STOK MINIMUM"))

		costPrice, _ := strconv.ParseInt(costPriceStr, 10, 64)
		basePrice, _ := strconv.ParseInt(sellPriceStr, 10, 64)
		currentStock, _ := strconv.Atoi(currentStockStr)
		minStock, _ := strconv.Atoi(minStockStr)
		
		isStockActive := true

		// Construct Input
		// Use getStringPtr for unique fields to avoid "duplicate key value violates unique constraint" on empty strings
		input := domain.ProductCreateInput{
			Name:          name,
			Barcode:       getStringPtr(barcode),
			SKU:           getStringPtr(sku),
			CategoryID:    categoryID,
			Unit:          unit,
			BasePrice:     basePrice,
			CostPrice:     costPrice,
			IsStockActive: &isStockActive,
			CurrentStock:  &currentStock,
			MinStockAlert: &minStock,
		}

		// Insert
		if _, err := productRepo.Create(ctx, input); err != nil {
			// Check if duplicate
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
				duplicateCount++
			} else {
				log.Printf("Failed to insert row %d (%s): %v", i+3, name, err)
				errorCount++
			}
		} else {
			successCount++
		}

		if (i+1)%2000 == 0 {
			log.Printf("Processed %d rows...", i+1)
		}
	}

	log.Printf("Seeding finished.\nSuccess: %d\nSkipped (Duplicates): %d\nStart Blocked (Name Empty): %d\nFailed (Other Errors): %d", successCount, duplicateCount, skippedCount, errorCount)
	return nil
}

func cleanNumber(s string) string {
	// Remove commas and dots
	s = strings.ReplaceAll(s, ",", "") 
	s = strings.ReplaceAll(s, ".", "")
	s = strings.TrimSpace(s)
	if s == "" {
		return "0"
	}
	// Extract only digits and leading minus if any
	// (Simple version: assuming clean input after replacing separators)
	return s
}
