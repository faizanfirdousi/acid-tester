// Package db handles the PostgreSQL connection.
// Uses database/sql with the lib/pq driver.
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // registers the "postgres" driver
)

// defaults match the credentials hardcoded in docker-compose.yml.
// Override any of these by setting environment variables before running.
const (
	defaultHost     = "localhost"
	defaultPort     = "5432"
	defaultUser     = "acid_user"
	defaultPassword = "acid_pass"
	defaultDBName   = "acid_db"
)

// env returns the value of an environment variable,
// or the fallback if it's not set. This means .env is optional —
// the binary works out of the box with just Docker running.
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Connect returns an open *sql.DB connection using defaults or env vars.
// *sql.DB is a connection pool — Go manages multiple connections automatically.
//
// Priority: environment variable > hardcoded default
func Connect() *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		env("DB_HOST", defaultHost),
		env("DB_PORT", defaultPort),
		env("DB_USER", defaultUser),
		env("DB_PASSWORD", defaultPassword),
		env("DB_NAME", defaultDBName),
	)

	// sql.Open validates the driver name but does NOT connect yet
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("❌ sql.Open failed: %v", err)
	}

	// db.Ping() actually dials the database
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Cannot reach Postgres — is the container running?\n   Run: docker compose up -d\n   Error: %v", err)
	}

	return db
}
