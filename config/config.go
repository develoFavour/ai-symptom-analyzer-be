package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port             string
	DatabaseURL      string
	JWTSecret        string
	JWTExpiry        time.Duration
	JWTRefreshSecret string
	JWTRefreshExpiry time.Duration
	GeminiAPIKey     string
	GroqAPIKey       string
	BrevoAPIKey      string
	BrevoSenderEmail string
	BrevoSenderName  string
	APIBaseURL       string
	FrontendURL      string
	DBPingInterval   time.Duration
	SelfPingInterval time.Duration
}

var App *AppConfig

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("[Config] No .env file found — using system environment variables")
	}

	App = &AppConfig{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      mustGetEnv("DATABASE_URL"),
		JWTSecret:        mustGetEnv("JWT_SECRET"),
		JWTExpiry:        parseDuration(getEnv("JWT_EXPIRY", "15m")),             // Reduced for better security, use refresh tokens
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", mustGetEnv("JWT_SECRET")), // Fallback to main secret if not set
		JWTRefreshExpiry: parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h")),    // 7 days
		GeminiAPIKey:     mustGetEnv("GEMINI_API_KEY"),
		GroqAPIKey:       mustGetEnv("GROQ_API_KEY"),
		BrevoAPIKey:      mustGetEnv("BREVO_API_KEY"),
		BrevoSenderEmail: mustGetEnv("BREVO_SENDER_EMAIL"),
		BrevoSenderName:  getEnv("BREVO_SENDER_NAME", "AI Symptom Checker"),
		APIBaseURL:       getEnv("API_BASE_URL", "http://localhost:8080"),
		FrontendURL:      getEnv("FRONTEND_URL", "http://localhost:3000"),
		DBPingInterval:   parseDuration(getEnv("DB_PING_INTERVAL", "120h")),  // every 5 days
		SelfPingInterval: parseDuration(getEnv("SELF_PING_INTERVAL", "10m")), // every 10 min
	}

	log.Println("[Config] Configuration loaded successfully")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("[Config] FATAL: Required environment variable '%s' is not set", key)
	}
	return val
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Fatalf("[Config] Invalid duration value '%s': %v", s, err)
	}
	return d
}
