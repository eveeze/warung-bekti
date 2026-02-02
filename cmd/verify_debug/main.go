package main

import (
	"context"
	"fmt"
	"log"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/database"
)

func main() {
    // 1. Check Database
	cfg := config.Load()
	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	var total, active, inactive int
	db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM products").Scan(&total)
	db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM products WHERE is_active = true").Scan(&active)
	db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM products WHERE is_active = false").Scan(&inactive)

	fmt.Println("--- DATABASE STATE ---")
	fmt.Printf("Total Products: %d\n", total)
	fmt.Printf("Active (True):  %d\n", active)
	fmt.Printf("Inactive (False): %d\n", inactive)
    fmt.Println("----------------------")
}
