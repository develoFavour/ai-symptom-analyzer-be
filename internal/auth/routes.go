package auth

import (
	"ai-symptom-checker/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	auth := r.Group("/auth")
	{
		auth.POST("/register", handler.Register)               // Patient self-registration
		auth.POST("/login", handler.Login)                     // Login for all roles
		auth.POST("/doctor/setup", handler.SetupDoctorAccount) // Complete setup via invite token
		auth.POST("/admin/setup", handler.SetupAdminAccount)   // Complete setup for invited admins
		auth.POST("/refresh", handler.Refresh)                 // Token renewal
		auth.GET("/verify-email", handler.VerifyEmail)         // Verify email via token

		// Protected routes
		auth.Use(middleware.AuthMiddleware())
		{
			auth.GET("/me", handler.GetMe)       // Get current user profile
			auth.POST("/logout", handler.Logout) // Clear session
			auth.POST("/doctor/wizard", handler.CompleteDoctorWizard)
		}
	}
}
