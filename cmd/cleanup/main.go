package main

import (
	"context"
	"fmt"
	"log"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/database"
)

func main() {
	cfg := config.Load()

	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Truncate all tables except schema_migrations
	query := `
		DO $$ DECLARE
			r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename != 'schema_migrations') LOOP
				EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE';
			END LOOP;
		END $$;
	`

	if _, err := db.ExecContext(ctx, query); err != nil {
		log.Fatalf("Failed to reset database: %v", err)
	}

	fmt.Println("Database reset successfully.")
}
