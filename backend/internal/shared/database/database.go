package database

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes a single GORM database connection instance.
// Defaults to SQLite unless DB_DRIVER=postgres is configured.
func InitDB() (*gorm.DB, error) {
	driver := os.Getenv("DB_DRIVER")
	dsn := os.Getenv("DB_DSN")

	var dialector gorm.Dialector

	if driver == "postgres" {
		if dsn == "" {
			dsn = "host=localhost user=postgres password=postgres dbname=inventory_db port=5432 sslmode=disable"
		}
		dialector = postgres.Open(dsn)
	} else {
		// Default SQLite
		if dsn == "" {
			dsn = "inventory.db"
		}
		dialector = sqlite.Open(dsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database (%s): %w", driver, err)
	}

	return db, nil
}
