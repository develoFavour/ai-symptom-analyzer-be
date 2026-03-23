package main

import (
	"log"

	"ai-symptom-checker/config"
	"ai-symptom-checker/pkg/keepalive"
	"ai-symptom-checker/router"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables & validate
	config.Load()

	// Connect to database
	db := config.ConnectDB()

	// Setup Gin engine
	r := gin.Default()

	// Register all routes
	router.Setup(r, db)

	// Start keep-alive goroutines (prevent Supabase + Render free-tier inactivity)
	keepalive.Start(keepalive.Config{
		DB:               db,
		DBPingInterval:   config.App.DBPingInterval,
		SelfPingInterval: config.App.SelfPingInterval,
		SelfURL:          config.App.APIBaseURL + "/health",
	})

	port := config.App.Port
	log.Printf("[Server] AI Symptom Checker API operational on port %s", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("[Server] Failed to start: %v", err)
	}
}
