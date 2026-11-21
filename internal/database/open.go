package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Open(storagePath string, logger *log.Logger) (*sql.DB, error) {
	dsn := "file:" + storagePath

	errorMkDir := os.MkdirAll(filepath.Dir(storagePath), 0755)

	if errorMkDir != nil {
		logger.Printf("failed to create the directory, error: %v", errorMkDir)
		return nil, errorMkDir
	}

	db, err := sql.Open("sqlite", dsn)

	if err != nil {
		logger.Printf("failed to open the database, error: %v", err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		logger.Printf("failed to ping the database, error: %v", err)
		// close the database connection if the database is not pinged
		defer db.Close()
		return nil, err
	}

	logger.Println("database opened successfully")

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		logger.Printf("failed to set foreign keys, error: %v", err)
		// close the database connection if the foreign keys are not set
		defer db.Close()
		return nil, err
	}

	return db, nil
}
