package router

import (
	"ai-symptom-checker/internal/admin"
	"ai-symptom-checker/internal/auth"
	"ai-symptom-checker/internal/consultation"
	"ai-symptom-checker/internal/doctor"
	"ai-symptom-checker/internal/knowledge"
	"ai-symptom-checker/internal/notification"
	"ai-symptom-checker/internal/report"
	"ai-symptom-checker/internal/symptom"
	"ai-symptom-checker/internal/user"
	"ai-symptom-checker/pkg/ai"
	"ai-symptom-checker/pkg/middleware"
	"ai-symptom-checker/pkg/socket"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(r *gin.Engine, db *gorm.DB) {
	// Global middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Health check — public, used by keep-alive goroutine & UptimeRobot
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "ai-symptom-checker-api",
			"timestamp": time.Now().UTC(),
		})
	})

	r.GET("/routes", func(c *gin.Context) {
		var routes []interface{}
		for _, info := range r.Routes() {
			routes = append(routes, map[string]string{
				"method": info.Method,
				"path":   info.Path,
			})
		}
		c.JSON(200, routes)
	})

	// Initialize shared AI client (Gemini primary, Groq fallback)
	aiClient := ai.NewResilientClient()

	// Initialize WebSocket Hub
	wsHub := socket.NewHub()
	go wsHub.Run()

	// ── API v1 ───────────────────────────────────────────────────────────────
	v1 := r.Group("/api/v1")

	// Public routes (no auth required)
	auth.RegisterRoutes(v1, db)

	// Symptom routes (self-contained auth inside the module)
	symptom.RegisterRoutes(v1, db, aiClient)

	// Protected routes (JWT required)
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		user.RegisterRoutes(protected, db)
		notifService := notification.RegisterRoutes(protected, db)
		consultation.RegisterRoutes(protected, db, aiClient, notifService, wsHub)
		knowledge.RegisterProtectedRoutes(protected, db, aiClient)

		// Doctor-only routes
		doctorGroup := protected.Group("")
		doctorGroup.Use(middleware.RoleMiddleware("doctor"))
		doctor.RegisterRoutes(doctorGroup, db)

		// Admin-only routes
		adminGroup := protected.Group("")
		adminGroup.Use(middleware.RoleMiddleware("admin"))
		admin.RegisterAdminRoutes(adminGroup, db)
		knowledge.RegisterAdminRoutes(adminGroup, db, aiClient)
		report.RegisterRoutes(adminGroup, db)
	}
}
