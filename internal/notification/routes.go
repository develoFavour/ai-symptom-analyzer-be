package notification

import (
	"ai-symptom-checker/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) *Service {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	notif := r.Group("/notifications")
	notif.Use(middleware.AuthMiddleware())
	{
		notif.GET("", handler.GetMyNotifications)
		notif.PATCH("/:id/read", handler.MarkAsRead)
	}

	return service
}
