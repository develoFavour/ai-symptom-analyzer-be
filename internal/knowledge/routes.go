package knowledge

import (
	"ai-symptom-checker/pkg/ai"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterAdminRoutes(r *gin.RouterGroup, db *gorm.DB, aiClient ai.Client) {
	repo := NewRepository(db)
	service := NewService(repo, aiClient)
	handler := NewHandler(service)

	admin := r.Group("/admin/knowledge")
	{
		admin.POST("/entries", handler.CreateEntry)
		admin.GET("/entries", handler.GetEntries)
		admin.GET("/alerts", handler.GetAlerts)
		admin.PUT("/entries/:id", handler.UpdateEntry)
		admin.DELETE("/entries/:id", handler.DeleteEntry)
	}
}

func RegisterProtectedRoutes(r *gin.RouterGroup, db *gorm.DB, aiClient ai.Client) {
	repo := NewRepository(db)
	service := NewService(repo, aiClient)
	handler := NewHandler(service)

	doctor := r.Group("/doctor/knowledge")
	{
		doctor.GET("/entries", handler.GetEntries)
	}
}
