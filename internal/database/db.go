package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// DB holds the database connection
type DB struct {
	*sql.DB
}

// NewConnection creates a new database connection
func NewConnection(dbPath string) (*DB, error) {
	// Create database file if it doesn't exist
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set optimal SQLite settings for web application
	pragmas := []string{
		"PRAGMA journal_mode=WAL",           // Write-Ahead Logging for concurrency
		"PRAGMA synchronous=NORMAL",         // Balance between safety and performance
		"PRAGMA cache_size=-64000",          // 64MB cache
		"PRAGMA foreign_keys=ON",            // Enable foreign key constraints
		"PRAGMA temp_store=MEMORY",          // Store temp tables in memory
		"PRAGMA busy_timeout=5000",          // 5 second timeout for locks
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("failed to set pragma %s: %w", pragma, err)
		}
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return &DB{db}, nil
}

// RunMigrations executes all SQL migration files
func (db *DB) RunMigrations(migrationsPath string) error {
	// Read all migration files
	files, err := ioutil.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Create migrations table to track applied migrations
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			filename TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Execute each migration file
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Check if migration has already been applied
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM migrations WHERE filename = ?)", file.Name()).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if exists {
			log.Printf("Skipping already applied migration: %s", file.Name())
			continue
		}

		// Read and execute migration file
		migrationPath := filepath.Join(migrationsPath, file.Name())
		content, err := ioutil.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		// Execute migration in a transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", file.Name(), err)
		}

		// Record migration as applied
		if _, err := tx.Exec("INSERT INTO migrations (filename) VALUES (?)", file.Name()); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", file.Name(), err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", file.Name(), err)
		}

		log.Printf("Applied migration: %s", file.Name())
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}