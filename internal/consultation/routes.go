package consultation

import (
	"ai-symptom-checker/internal/notification"
	"ai-symptom-checker/pkg/ai"
	"ai-symptom-checker/pkg/middleware"
	"ai-symptom-checker/pkg/socket"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, aiClient ai.Client, notifSvc *notification.Service, hub *socket.Hub) {
	repo := NewRepository(db)
	svc := NewService(repo, aiClient, notifSvc)
	handler := NewHandler(svc, hub)

	consultRoutes := r.Group("/consultations")
	consultRoutes.Use(middleware.AuthMiddleware())
	{
		consultRoutes.GET("/doctors", handler.ListDoctors) // Browse available doctors
		consultRoutes.POST("", handler.CreateConsultation) // Submit a consultation request
		consultRoutes.GET("", handler.GetMyConsultations)  // Patient: view own requests
		consultRoutes.GET("/:id", handler.GetConsultation) // View details (Patient or Assigned Doctor)
		consultRoutes.POST("/:id/messages", handler.AddMessage)
		consultRoutes.GET("/:id/ws", handler.WebSocketUpgrade)
	}

	// Doctor-only routes
	doctorRoutes := r.Group("/doctor/consultations")
	doctorRoutes.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("doctor"))
	{
		doctorRoutes.GET("", handler.ListDoctorQueue)
		doctorRoutes.POST("/:id/reply", handler.ReplyToConsultation)
	}
}
