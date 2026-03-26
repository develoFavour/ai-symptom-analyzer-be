package config

import (
	"log"
	"os"
	"time"

	"ai-symptom-checker/models"

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

	// Auto-migrate only if enabled
	if os.Getenv("SKIP_MIGRATIONS") == "true" {
		log.Println("[Database] Skipping auto-migrations (SKIP_MIGRATIONS=true)")
	} else {
		log.Println("[Database] Running auto-migrations...")
		mStart := time.Now()
		err = db.AutoMigrate(
			&models.User{},
			&models.Doctor{},
			&models.Admin{},
			&models.SymptomSession{},
			&models.Diagnosis{},
			&models.Consultation{},
			&models.ConsultationReply{},
			&models.ConsultationMessage{},
			&models.KnowledgeEntry{},
			&models.Notification{},
			&models.Feedback{},
			&models.AdminLog{},
		)
		if err != nil {
			log.Fatalf("[Database] Auto-migration failed: %v", err)
		}
		log.Printf("[Database] Auto-migration completed in %v", time.Since(mStart))
	}

	DB = db
	return db
}
