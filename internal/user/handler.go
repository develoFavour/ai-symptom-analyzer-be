package user

import (
	"ai-symptom-checker/pkg/middleware"
	"ai-symptom-checker/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler holds service dependency for the user module
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetDashboard(c *gin.Context) {
	userID, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		utils.Unauthorized(c, "authentication required")
		return
	}

	data, err := h.service.GetDashboardData(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		utils.InternalServerError(c, "failed to load dashboard", err)
		return
	}

	utils.Ok(c, "success", data)
}
