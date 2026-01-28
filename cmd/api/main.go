package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/pkg/logger"
	"github.com/eveeze/warung-backend/internal/router"
	"github.com/eveeze/warung-backend/internal/storage"
)

func main() {
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

	// Run migrations
	migrator := database.NewMigrator(db)
	if err := migrator.Up(context.Background()); err != nil {
		logger.Fatal("Failed to run migrations: %v", err)
	}

	// Connect to Redis
	redis, err := database.NewRedis(&cfg.Redis)
	if err != nil {
		logger.Warn("Failed to connect to Redis: %v", err)
		redis = nil // Continue without Redis
	} else {
		defer redis.Close()
	}

	// Connect to Minio
	minio, err := storage.NewMinio(&cfg.Minio)
	if err != nil {
		logger.Warn("Failed to connect to Minio: %v", err)
		minio = nil // Continue without Minio
	}

	// Setup router
	handler := router.New(cfg, db, redis, minio)

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
