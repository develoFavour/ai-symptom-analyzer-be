package symptom

import (
	"ai-symptom-checker/pkg/ai"
	"ai-symptom-checker/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, aiClient ai.Client) {
	repo := NewRepository(db)
	svc := NewService(repo, aiClient)
	handler := NewHandler(svc)

	// All symptom routes REQUIRE authentication
	symptomRoutes := r.Group("/symptoms")
	symptomRoutes.Use(middleware.AuthMiddleware())
	{
		// Session management
		symptomRoutes.POST("/sessions", handler.StartSession)           // Create a new session → returns server ID
		symptomRoutes.GET("/sessions", handler.ListSessions)            // List all patient's sessions
		symptomRoutes.GET("/sessions/:sessionId", handler.GetSession)   // Load a session (for page refresh)

		// Chat
		symptomRoutes.POST("/chat", handler.HandleChat)                 // Send a chat message to the AI
	}
}
