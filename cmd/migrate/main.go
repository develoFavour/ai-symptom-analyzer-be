package main

import (
	"log"
	"time"

	"ai-symptom-checker/config"
	"ai-symptom-checker/models"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or failed to load, proceeding with environment variables")
	}

	config.Load()
	db := config.ConnectDB()

	log.Println("[Database] Running migrations...")
	start := time.Now()
	if err := db.AutoMigrate(
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
	); err != nil {
		log.Fatalf("[Database] Migration failed: %v", err)
	}

	log.Printf("[Database] Migrations completed in %v", time.Since(start))
}
