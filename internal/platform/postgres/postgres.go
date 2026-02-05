package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/souravkumar/distributed-rate-limiter/internal/platform/config"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// New creates a new GORM database connection and runs migrations
func New(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.Postgres.GetDSN()

	// First, run migrations using database/sql
	if err := runMigrations(dsn, cfg.Postgres.Database); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Silent)
	if cfg.Logging.Level == "debug" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Connect with GORM for ORM operations
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying db: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Postgres.MaxConnections)
	sqlDB.SetMaxIdleConns(cfg.Postgres.MaxConnections / 2)

	log.Info().Msg("database connected and migrated")
	return db, nil
}

// runMigrations runs all pending database migrations
func runMigrations(dsn, dbName string) error {
	log.Info().Msg("running database migrations")

	// Open connection for migrations
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	// Create postgres driver for migrate
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		dbName,
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, dirty, _ := m.Version()
	log.Info().Uint("version", version).Bool("dirty", dirty).Msg("database migrations completed")

	return nil
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
