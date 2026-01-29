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

	// Check if admin exists
	adminEmail := "admin@warung.com"
	existingAdmin, err := userRepo.GetByEmail(ctx, adminEmail)
	if err == nil && existingAdmin != nil {
		log.Println("Admin user already exists. Skipping seeding.")
		return
	}

	// Create admin user
	log.Println("Seeding admin user...")
	
	defaultPassword := "password" // In production this should be env var or random
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

	if err := userRepo.Create(ctx, admin); err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	log.Printf("Admin user created successfully.\nEmail: %s\nPassword: %s\n", adminEmail, defaultPassword)
}
