package keepalive

import (
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// Config holds keep-alive configuration
type Config struct {
	DB               *gorm.DB
	DBPingInterval   time.Duration // Recommended: every 5 days (Supabase pauses at 7)
	SelfPingInterval time.Duration // Recommended: every 10 min (Render spins down at 15)
	SelfURL          string        // e.g. https://your-api.onrender.com/health
}

// Start launches all keep-alive goroutines — call this after server is configured
func Start(cfg Config) {
	go keepDatabaseAlive(cfg.DB, cfg.DBPingInterval)
	go keepServerAlive(cfg.SelfURL, cfg.SelfPingInterval)
	log.Println("[KeepAlive] Keep-alive goroutines started")
}

// keepDatabaseAlive pings Supabase PostgreSQL every N interval
// to prevent the free-tier project from auto-pausing
func keepDatabaseAlive(db *gorm.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("[KeepAlive] DB ping scheduled every %s", interval)

	for range ticker.C {
		var result int
		err := db.Raw("SELECT 1").Scan(&result).Error
		if err != nil {
			log.Printf("[KeepAlive] DB ping FAILED: %v", err)
		} else {
			log.Println("[KeepAlive] DB ping OK — Supabase activity registered")
		}
	}
}

// keepServerAlive pings the /health endpoint to prevent
// Render.com from spinning down the free-tier service
func keepServerAlive(selfURL string, interval time.Duration) {
	// Wait 30s for the server to fully start before pinging itself
	time.Sleep(30 * time.Second)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("[KeepAlive] Self-ping scheduled every %s → %s", interval, selfURL)

	client := &http.Client{Timeout: 10 * time.Second}

	for range ticker.C {
		resp, err := client.Get(selfURL)
		if err != nil {
			log.Printf("[KeepAlive] Self-ping FAILED: %v", err)
			continue
		}
		resp.Body.Close()
		log.Printf("[KeepAlive] Self-ping OK — status %d", resp.StatusCode)
	}
}
