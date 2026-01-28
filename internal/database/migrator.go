package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/eveeze/warung-backend/internal/pkg/logger"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migration represents a database migration
type Migration struct {
	Version  string
	Name     string
	UpSQL    string
	DownSQL  string
	Applied  bool
	AppliedAt *time.Time
}

// Migrator handles database migrations
type Migrator struct {
	db *PostgresDB
}

// NewMigrator creates a new Migrator
func NewMigrator(db *PostgresDB) *Migrator {
	return &Migrator{db: db}
}

// Init creates the migrations tracking table
func (m *Migrator) Init(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	return nil
}

// LoadMigrations loads all migrations from the embedded filesystem
func (m *Migrator) LoadMigrations() ([]Migration, error) {
	migrations := make(map[string]*Migration)

	err := fs.WalkDir(migrationsFS, "migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		filename := filepath.Base(path)
		if !strings.HasSuffix(filename, ".sql") {
			return nil
		}

		// Parse filename: 001_init_schema.up.sql or 001_init_schema.down.sql
		parts := strings.Split(filename, ".")
		if len(parts) < 3 {
			return nil
		}

		direction := parts[len(parts)-2] // "up" or "down"
		baseName := strings.Join(parts[:len(parts)-2], ".")
		
		// Extract version (first part before underscore)
		nameParts := strings.SplitN(baseName, "_", 2)
		version := nameParts[0]
		name := baseName

		content, err := fs.ReadFile(migrationsFS, path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		if migrations[version] == nil {
			migrations[version] = &Migration{
				Version: version,
				Name:    name,
			}
		}

		switch direction {
		case "up":
			migrations[version].UpSQL = string(content)
		case "down":
			migrations[version].DownSQL = string(content)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert map to sorted slice
	var result []Migration
	for _, migration := range migrations {
		result = append(result, *migration)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	return result, nil
}

// LoadMigrationsFromDir loads migrations from a specific directory path
func (m *Migrator) LoadMigrationsFromDir(dir string) ([]Migration, error) {
	migrations := make(map[string]*Migration)

	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, path := range files {
		filename := filepath.Base(path)
		parts := strings.Split(filename, ".")
		if len(parts) < 3 {
			continue
		}

		direction := parts[len(parts)-2]
		baseName := strings.Join(parts[:len(parts)-2], ".")
		nameParts := strings.SplitN(baseName, "_", 2)
		version := nameParts[0]
		name := baseName

		content, err := fs.ReadFile(nil, path)
		if err != nil {
			// Fallback to os.ReadFile for external files
			continue
		}

		if migrations[version] == nil {
			migrations[version] = &Migration{
				Version: version,
				Name:    name,
			}
		}

		switch direction {
		case "up":
			migrations[version].UpSQL = string(content)
		case "down":
			migrations[version].DownSQL = string(content)
		}
	}

	var result []Migration
	for _, migration := range migrations {
		result = append(result, *migration)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	return result, nil
}

// GetAppliedMigrations returns a list of applied migration versions
func (m *Migrator) GetAppliedMigrations(ctx context.Context) (map[string]time.Time, error) {
	query := `SELECT version, applied_at FROM schema_migrations ORDER BY version`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]time.Time)
	for rows.Next() {
		var version string
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, err
		}
		applied[version] = appliedAt
	}

	return applied, rows.Err()
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context) error {
	if err := m.Init(ctx); err != nil {
		return err
	}

	migrations, err := m.LoadMigrations()
	if err != nil {
		return err
	}

	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if _, ok := applied[migration.Version]; ok {
			logger.Debug("Migration %s already applied, skipping", migration.Name)
			continue
		}

		logger.Info("Applying migration: %s", migration.Name)

		if err := m.runMigration(ctx, migration.Version, migration.UpSQL); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Name, err)
		}

		logger.Info("Migration %s applied successfully", migration.Name)
	}

	return nil
}

// Down rolls back the last migration
func (m *Migrator) Down(ctx context.Context) error {
	migrations, err := m.LoadMigrations()
	if err != nil {
		return err
	}

	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	// Find the last applied migration
	var lastMigration *Migration
	for i := len(migrations) - 1; i >= 0; i-- {
		if _, ok := applied[migrations[i].Version]; ok {
			lastMigration = &migrations[i]
			break
		}
	}

	if lastMigration == nil {
		logger.Info("No migrations to rollback")
		return nil
	}

	logger.Info("Rolling back migration: %s", lastMigration.Name)

	if err := m.rollbackMigration(ctx, lastMigration.Version, lastMigration.DownSQL); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", lastMigration.Name, err)
	}

	logger.Info("Migration %s rolled back successfully", lastMigration.Name)

	return nil
}

// Reset rolls back all migrations
func (m *Migrator) Reset(ctx context.Context) error {
	for {
		applied, err := m.GetAppliedMigrations(ctx)
		if err != nil {
			return err
		}
		if len(applied) == 0 {
			break
		}
		if err := m.Down(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Status prints the migration status
func (m *Migrator) Status(ctx context.Context) ([]Migration, error) {
	if err := m.Init(ctx); err != nil {
		return nil, err
	}

	migrations, err := m.LoadMigrations()
	if err != nil {
		return nil, err
	}

	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	for i := range migrations {
		if appliedAt, ok := applied[migrations[i].Version]; ok {
			migrations[i].Applied = true
			migrations[i].AppliedAt = &appliedAt
		}
	}

	return migrations, nil
}

func (m *Migrator) runMigration(ctx context.Context, version, sql string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, sql); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, 
		`INSERT INTO schema_migrations (version) VALUES ($1)`, 
		version,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Migrator) rollbackMigration(ctx context.Context, version, sql string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, sql); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, 
		`DELETE FROM schema_migrations WHERE version = $1`, 
		version,
	); err != nil {
		return err
	}

	return tx.Commit()
}

// MigrateUp is a convenience function to run migrations
func MigrateUp(db *PostgresDB) error {
	migrator := NewMigrator(db)
	return migrator.Up(context.Background())
}

// MigrateDown is a convenience function to rollback the last migration
func MigrateDown(db *PostgresDB) error {
	migrator := NewMigrator(db)
	return migrator.Down(context.Background())
}

// Ensure PostgresDB implements the needed interface
var _ interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
} = (*PostgresDB)(nil)
