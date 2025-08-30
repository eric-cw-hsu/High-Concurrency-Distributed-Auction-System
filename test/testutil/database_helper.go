package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
)

// DatabaseTestHelper provides common database testing utilities
type DatabaseTestHelper struct {
	DB     *sql.DB
	Config *config.PostgresConfig
	ctx    context.Context
}

// NewDatabaseTestHelper creates a new database test helper
func NewDatabaseTestHelper(ctx context.Context) (*DatabaseTestHelper, error) {
	// Load test environment
	if err := loadTestEnv(); err != nil {
		return nil, fmt.Errorf("failed to load test environment: %w", err)
	}

	// Get database configuration
	pgConfig := getPostgresConfig()

	// Connect to database
	dsn := buildDSN(pgConfig)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	helper := &DatabaseTestHelper{
		DB:     db,
		Config: pgConfig,
		ctx:    ctx,
	}

	return helper, nil
}

// RunMigrations executes database migrations
func (h *DatabaseTestHelper) RunMigrations() error {
	if err := h.dropTables(); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	// Get project root directory
	projectRoot, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to get project root: %w", err)
	}

	migrationsPath := filepath.Join(projectRoot, "db", "migrations")
	sourceURL := fmt.Sprintf("file://%s", migrationsPath)
	migrationDSN := buildMigrationDSN(h.Config)

	m, err := migrate.New(sourceURL, migrationDSN)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// CleanDatabase cleans all test data from database
func (h *DatabaseTestHelper) CleanDatabase() error {
	// Get all table names from the database
	query := `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename NOT LIKE 'schema_migrations%'
	`

	rows, err := h.DB.QueryContext(h.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over table names: %w", err)
	}

	// Truncate all tables
	for _, table := range tables {
		_, err := h.DB.ExecContext(h.ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	return nil
}

// dropTables drops all tables in the database
func (h *DatabaseTestHelper) dropTables() error {
	// Get all table names from the database
	query := `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public'
	`

	rows, err := h.DB.QueryContext(h.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over table names: %w", err)
	}

	// Drop all tables
	for _, table := range tables {
		_, err := h.DB.ExecContext(h.ctx, fmt.Sprintf("DROP TABLE %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}

// Close closes the database connection
func (h *DatabaseTestHelper) Close() error {
	if h.DB != nil {
		return h.DB.Close()
	}
	return nil
}

// buildDSN builds a PostgreSQL DSN string
func buildDSN(config *config.PostgresConfig) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)
}

// buildMigrationDSN builds a PostgreSQL DSN string for migrations
func buildMigrationDSN(config *config.PostgresConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.User, config.Password, config.Host, config.Port, config.DBName)
}

// loadTestEnv loads the test environment file
func loadTestEnv() error {
	projectRoot, err := getProjectRoot()
	if err != nil {
		return err
	}

	envFile := filepath.Join(projectRoot, ".env.test")
	if _, err := os.Stat(envFile); err == nil {
		return godotenv.Load(envFile)
	}

	return fmt.Errorf("test environment file not found: %s", envFile)
}

// getProjectRoot returns the project root directory
func getProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Navigate up to find project root (contains go.mod)
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}

	return "", fmt.Errorf("project root not found")
}

// getPostgresConfig creates postgres config from environment
func getPostgresConfig() *config.PostgresConfig {
	port, _ := strconv.Atoi(getEnvOrDefault("POSTGRES_PORT", "5432"))

	return &config.PostgresConfig{
		Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:     port,
		User:     getEnvOrDefault("POSTGRES_USER", "postgres"),
		Password: getEnvOrDefault("POSTGRES_PASSWORD", "password"),
		DBName:   getEnvOrDefault("POSTGRES_DB", "auction_db_test"),
	}
}

// getEnvOrDefault gets environment variable or returns default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
