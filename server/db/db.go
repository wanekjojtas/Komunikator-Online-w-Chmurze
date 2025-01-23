package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

// NewDatabase initializes a new database connection
func NewDatabase() (*Database, error) {
	// Attempt to get the database URL from the environment variable
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("DATABASE_URL is not set. Falling back to local database.")
		// Use local PostgreSQL for development
		dsn = "postgresql://root:password@localhost:5433/go-chat?sslmode=disable"
	}

	// Open the database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to the database")
	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() {
	if err := d.db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	} else {
		log.Println("Database connection closed successfully")
	}
}

// GetDB returns the underlying sql.DB instance
func (d *Database) GetDB() *sql.DB {
	return d.db
}
