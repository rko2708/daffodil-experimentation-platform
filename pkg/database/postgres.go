package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // Postgres driver
)

// DBConfig holds connection details
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// NewPostgresConn creates a validated connection to Postgres
func NewPostgresConn(cfg DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Verify the connection is actually alive
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Printf("Successfully connected to Postgres at %s:%d", cfg.Host, cfg.Port)
	return db, nil
}
