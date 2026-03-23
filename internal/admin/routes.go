package admin

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterAdminRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	admin := r.Group("/admin")
	{
		// Administrative Onboarding & Team Management (Moved to Top)
		admin.POST("/invite", handler.InviteAdmin)
		admin.GET("/admins", handler.ListAdmins)
		admin.DELETE("/admins/:id", handler.DeleteAdmin)

		// Stats
		admin.GET("/stats", handler.GetStats)

		// Doctor Management
		admin.POST("/doctors/invite", handler.InviteDoctor)
		admin.GET("/doctors", handler.ListDoctors)
		admin.PATCH("/doctors/:id/status", handler.UpdateDoctorStatus)

		// Patient/User Management
		admin.GET("/users", handler.ListUsers)
		admin.PATCH("/users/:id/status", handler.UpdateUserStatus)
		admin.DELETE("/users/:id", handler.DeleteUser)
	}
}
