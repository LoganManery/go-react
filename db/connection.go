package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// DBConfig holds database connection configuration
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Database represents a database connection pool
type Database struct {
	Pool *pgxpool.Pool
}

// NewDBConfig creates a new database configuration with default values
func NewDBConfig() DBConfig {
	return DBConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "web_application_db",
		SSLMode:  "disable",
	}
}

// Connect establishes a connection to the database
func Connect(config DBConfig) (*Database, error) {
	connString := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=%s&pool_max_conns=10&pool_max_conn_lifetime=1h",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		config.SSLMode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to the database")
	return &Database{Pool: pool}, nil
}

// Close closes the database connection pool
func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// InTransaction executes a function within a transaction
func (db *Database) InTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	// Begin transaction
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	// Execute the function
	err = fn(tx)

	// Handle the result
	if err != nil {
		// Rollback if there was an error
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("error rolling back transaction: %w (original error: %v)", rbErr, err)
		}
		return err
	}

	// Commit if everything was successful
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
