package config

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() *gorm.DB {
	dsn := App.DatabaseURL

	start := time.Now()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Warn), // Show warnings/slow queries
		PrepareStmt: false,
	})
	if err != nil {
		log.Fatalf("[Database] Failed to connect to PostgreSQL: %v", err)
	}
	log.Printf("[Database] Connection established in %v", time.Since(start))

	// Set connection pool limits
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	DB = db
	return db
}
