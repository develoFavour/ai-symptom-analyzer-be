package config

import (
	"log"

	"ai-symptom-checker/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() *gorm.DB {
	dsn := App.DatabaseURL

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Silent),
		PrepareStmt: false, // Fix "prepared statement name is already in use" (08P01)
	})
	if err != nil {
		log.Fatalf("[Database] Failed to connect to PostgreSQL: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Doctor{},
		&models.Admin{},
		&models.SymptomSession{},
		&models.Diagnosis{},
		&models.Consultation{},
		&models.ConsultationReply{},
		&models.ConsultationMessage{}, // Added missing model
		&models.KnowledgeEntry{},
		&models.Notification{},
		&models.Feedback{},
		&models.AdminLog{},
	)
	if err != nil {
		log.Fatalf("[Database] Auto-migration failed: %v", err)
	}

	DB = db
	log.Println("[Database] Connected to PostgreSQL database successfully")
	return db
}
