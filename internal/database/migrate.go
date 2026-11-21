package database

import (
	"database/sql"
	"log"
	"os"
	"time"
)

func Migrate(db *sql.DB, logger *log.Logger) error {

	const smokeMigration = "0001_smoke.up.sql"

	const createSchemaMigrationsTable = `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TEXT NOT NULL
	)
	`

	_, err := db.Exec(createSchemaMigrationsTable)
	if err != nil {
		logger.Printf("failed to create the schema_migrations table, error: %v", err)
		return err
	}

	query, err := os.ReadFile("migrations/" + smokeMigration)
	if err != nil {
		logger.Printf("failed to read the migration file, error: %v", err)
		return err
	}

	transaction, err := db.Begin()
	if err != nil {
		transaction.Rollback()
		logger.Printf("failed to begin the transaction, error: %v", err)
		return err
	}

	_, err = transaction.Exec(string(query))
	if err != nil {
		logger.Printf("failed to execute the migration file, error: %v", err)
		return err
	}

	transaction.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)", smokeMigration, time.Now().Format(time.RFC3339))

	transaction.Commit()

	return nil
}
