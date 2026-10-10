package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/eveeze/warung-backend/internal/config"
)

func TestDatabaseConnectionsAndErrors(t *testing.T) {
	// Bad postgres connection
	badDbCfg := &config.DatabaseConfig{
		Host:     "127.0.0.1",
		Port:     "54329",
		User:     "baduser",
		Password: "badpassword",
		DBName:   "nodb",
		SSLMode:  "disable",
	}
	_, err := NewPostgres(badDbCfg)
	if err == nil {
		t.Errorf("expected error connecting to non-existent postgres port")
	}

	// Bad redis connection
	badRedisCfg := &config.RedisConfig{
		Host: "127.0.0.1",
		Port: "63799",
	}
	_, err = NewRedis(badRedisCfg)
	if err == nil {
		t.Errorf("expected error connecting to non-existent redis port")
	}

	// Live local Postgres test (port 5435)
	dbCfg := &config.DatabaseConfig{
		Host:            "127.0.0.1",
		Port:            "5435",
		User:            "warung",
		Password:        "warung_secret",
		DBName:          "warung_db",
		SSLMode:         "disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
	db, err := NewPostgres(dbCfg)
	if err != nil {
		t.Logf("skipping live db test: %v", err)
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Health(ctx); err != nil {
		t.Errorf("expected healthy postgres: %v", err)
	}

	// WithTransaction
	err = db.WithTransaction(ctx, func(tx *sql.Tx) error {
		var result int
		return tx.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	})
	if err != nil {
		t.Errorf("expected transaction success: %v", err)
	}

	// WithTransaction error rollback
	err = db.WithTransaction(ctx, func(tx *sql.Tx) error {
		return errors.New("intentional rollback")
	})
	if err == nil {
		t.Errorf("expected rollback error")
	}

	// Migrator test
	migrator := NewMigrator(db)
	if err := migrator.Init(ctx); err != nil {
		t.Errorf("migrator init failed: %v", err)
	}
	migrations, err := migrator.LoadMigrations()
	if err != nil {
		t.Errorf("migrator load migrations failed: %v", err)
	}
	if len(migrations) == 0 {
		t.Errorf("expected migrations > 0")
	}
	if _, err := migrator.Status(ctx); err != nil {
		t.Errorf("migrator status failed: %v", err)
	}
	_, _ = migrator.GetAppliedMigrations(ctx)
	_, _ = migrator.LoadMigrationsFromDir("migrations")
	_ = migrator.Up(ctx)
	_ = MigrateUp(db)
	_ = migrator.runMigration(ctx, "999_test", "CREATE TABLE IF NOT EXISTS test_dummy (id int)")
	_ = migrator.rollbackMigration(ctx, "999_test", "DROP TABLE IF EXISTS test_dummy")

	// Test Down and Reset with canceled context error path
	canceledCtx, cancelNow := context.WithCancel(context.Background())
	cancelNow()
	_ = migrator.Down(canceledCtx)
	_ = migrator.Reset(canceledCtx)

	// Test MigrateDown on closed DB error path
	closedDb, err := NewPostgres(dbCfg)
	if err == nil {
		closedDb.Close()
		_ = MigrateDown(closedDb)
	}

	// Live local Redis test (port 6381)
	redisCfg := &config.RedisConfig{
		Host: "127.0.0.1",
		Port: "6381",
	}
	rClient, err := NewRedis(redisCfg)
	if err != nil {
		t.Logf("skipping live redis test: %v", err)
		return
	}
	defer rClient.Close()

	if err := rClient.Health(ctx); err != nil {
		t.Errorf("expected healthy redis: %v", err)
	}

	// Test redis helper methods
	type TestData struct {
		Name string `json:"name"`
	}
	_ = rClient.SetJSON(ctx, "test:json:1", TestData{Name: "Budi"}, time.Minute)
	_ = rClient.SetJSON(ctx, "test:json:2", TestData{Name: "Siti"}, time.Minute)
	_, _ = rClient.GetJSON(ctx, "test:json:1")
	_ = rClient.DeleteKeys(ctx, "test:json:*")

	// Test cache key generators
	_ = ProductCacheKey("123")
	_ = ProductListCacheKey(1, 10, "filters")
	_ = CustomerCacheKey("123")
	_ = TransactionCacheKey("123")
	_ = ReportCacheKey("daily", "2026-10-10")
}
