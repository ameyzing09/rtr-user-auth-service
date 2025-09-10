package config

import (
	"fmt"
	"log"

	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/entities"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database holds the database connection
type Database struct {
	DB *gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(config *DatabaseConfig) (*Database, error) {
	dsn := config.GetDSN()
	
	// Configure GORM logger
	var gormLogger logger.Interface
	gormLogger = logger.Default.LogMode(logger.Info)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return &Database{DB: db}, nil
}

// Migrate runs database migrations
func (d *Database) Migrate() error {
	log.Println("Running database migrations...")
	
	// Auto-migrate all entities
	err := d.DB.AutoMigrate(
		&entities.Tenant{},
		&entities.User{},
		&entities.RefreshToken{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping checks if the database connection is alive
func (d *Database) Ping() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}