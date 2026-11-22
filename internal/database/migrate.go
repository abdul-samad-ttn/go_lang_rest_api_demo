package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func migrationApplied(db *sql.DB, version string) (bool, error) {
	var one int
	err := db.QueryRow("SELECT 1 FROM schema_migrations WHERE version = ?", version).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func applyMigration(db *sql.DB, logger *log.Logger, version string) error {
	query, err := os.ReadFile(filepath.Join("migrations", version))
	if err != nil {
		logger.Printf("failed to read the migration file, error: %v", err)
		return err
	}

	applied, err := migrationApplied(db, version)
	if err != nil {
		return err
	}

	if applied {
		logger.Printf("migration %s already applied", version)
		return nil
	}

	transaction, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = transaction.Exec(string(query))
	if err != nil {
		transaction.Rollback()
		return err
	}

	_, err = transaction.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)", version, time.Now().Format(time.RFC3339))
	if err != nil {
		transaction.Rollback()
		return err
	}

	err = transaction.Commit()
	if err != nil {
		return err
	}

	logger.Printf("migration %s applied successfully", version)
	return nil
}

func Migrate(db *sql.DB, logger *log.Logger) error {

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

	entries, err := os.ReadDir("migrations")
	if err != nil {
		logger.Printf("failed to read the migrations directory, error: %v", err)
		return err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".sql") {
			files = append(files, name)
		}
	}
	sort.Strings(files)

	for _, version := range files {
		if err := applyMigration(db, logger, version); err != nil {
			logger.Printf("failed to apply migration %s, error: %v", version, err)
			return err
		}
	}

	return nil
}
