package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/pkg/logger"
)

// PostgresDB wraps the sql.DB with additional functionality
type PostgresDB struct {
	*sql.DB
	config *config.DatabaseConfig
}

// NewPostgres creates a new PostgreSQL connection
func NewPostgres(cfg *config.DatabaseConfig) (*PostgresDB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("PostgreSQL connection established successfully")

	return &PostgresDB{
		DB:     db,
		config: cfg,
	}, nil
}

// Close closes the database connection
func (db *PostgresDB) Close() error {
	logger.Info("Closing PostgreSQL connection")
	return db.DB.Close()
}

// Health checks the database health
func (db *PostgresDB) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}

// BeginTx starts a new transaction with the given options
func (db *PostgresDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.DB.BeginTx(ctx, opts)
}

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func (db *PostgresDB) WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// QueryRowContext is a wrapper that adds logging in debug mode
func (db *PostgresDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	logger.Default().WithFields(map[string]interface{}{
		"query": query,
		"args":  args,
	}).Debug("Executing query row")
	return db.DB.QueryRowContext(ctx, query, args...)
}

// QueryContext is a wrapper that adds logging in debug mode
func (db *PostgresDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	logger.Default().WithFields(map[string]interface{}{
		"query": query,
		"args":  args,
	}).Debug("Executing query")
	return db.DB.QueryContext(ctx, query, args...)
}

// ExecContext is a wrapper that adds logging in debug mode
func (db *PostgresDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	logger.Default().WithFields(map[string]interface{}{
		"query": query,
		"args":  args,
	}).Debug("Executing statement")
	return db.DB.ExecContext(ctx, query, args...)
}
