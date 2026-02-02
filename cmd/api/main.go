package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/pkg/logger"
	"github.com/eveeze/warung-backend/internal/router"
	"github.com/eveeze/warung-backend/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	_ = godotenv.Load()

	// Simple CLI parsing
	if len(os.Args) < 2 {
		// Default to running the server
		runServer()
		return
	}

	command := os.Args[1]

	switch command {
	case "server":
		runServer()
	case "migrate":
		handleMigrate(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: warung-api [server|migrate]")
		os.Exit(1)
	}
}

func handleMigrate(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: migrate [up|down|status]")
		os.Exit(1)
	}

	cfg := config.Load()
	
	// Setup logger for migration output
	log := logger.New(os.Stdout, logger.LevelInfo)
	logger.SetDefault(log)

	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	migrator := database.NewMigrator(db)
	ctx := context.Background()

	subCmd := args[0]
	switch subCmd {
	case "up":
		if err := migrator.Up(ctx); err != nil {
			logger.Fatal("Migration up failed: %v", err)
		}
		logger.Info("Migrations up completed successfully")
	case "down":
		if err := migrator.Down(ctx); err != nil {
			logger.Fatal("Migration down failed: %v", err)
		}
		logger.Info("Migration down completed successfully")
	case "status":
		migrations, err := migrator.Status(ctx)
		if err != nil {
			logger.Fatal("Failed to get migration status: %v", err)
		}
		for _, m := range migrations {
			status := "Pending"
			if m.Applied {
				status = fmt.Sprintf("Applied at %s", m.AppliedAt.Format(time.RFC3339))
			}
			fmt.Printf("[%s] %s - %s\n", m.Version, m.Name, status)
		}
	default:
		fmt.Printf("Unknown migration command: %s\n", subCmd)
		os.Exit(1)
	}
}

func runServer() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	log := logger.New(os.Stdout, logger.ParseLevel(cfg.App.LogLevel))
	logger.SetDefault(log)

	logger.Info("Starting WarungOS Backend API")
	logger.Info("Environment: %s", cfg.App.Environment)

	// Connect to PostgreSQL
	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Run migrations ON STARTUP only if configured or default behavior?
	// For production stability, usually we want explicit migration.
	// But to keep existing behavior for "server" command (except the conflict), we can skip it here 
	// or make it optional. 
	// The problem was `runServer` causes port bind. `migrate` command shouldn't run `runServer`.
	// So `runServer` CAN run migrations if we want auto-migrate on start.
	// But `make migrate-up` calls `go run main.go migrate up`, which now won't call `runServer`.
	// So we are safe.
	
	// Optional: Auto-migrate on startup (can be disabled)
	// migrator := database.NewMigrator(db)
	// if err := migrator.Up(context.Background()); err != nil {
	// 	logger.Fatal("Failed to run migrations: %v", err)
	// }

	// Connect to Redis
	redis, err := database.NewRedis(&cfg.Redis)
	if err != nil {
		logger.Warn("Failed to connect to Redis: %v", err)
		redis = nil // Continue without Redis
	} else {
		defer redis.Close()
	}

	// Connect to R2
	r2, err := storage.NewR2(&cfg.R2)
	if err != nil {
		logger.Warn("Failed to connect to R2: %v", err)
		r2 = nil // Continue without storage
	}

	// Setup router
	handler := router.New(cfg, db, redis, r2)

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("HTTP server listening on %s", cfg.Server.Address())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited properly")
}
