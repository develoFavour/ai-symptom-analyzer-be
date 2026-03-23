package notification

import (
	"ai-symptom-checker/pkg/middleware"
	"ai-symptom-checker/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetMyNotifications(c *gin.Context) {
	userIDRaw, _ := c.Get(middleware.ContextKeyUserID)
	userID, ok := userIDRaw.(uuid.UUID)
	if !ok {
		utils.Unauthorized(c, "User not found in context")
		return
	}

	notifications, err := h.service.GetUserNotifications(c.Request.Context(), userID)
	if err != nil {
		utils.InternalServerError(c, "Failed to fetch notifications", err)
		return
	}

	utils.Ok(c, "Notifications retrieved", notifications)
}

func (h *Handler) MarkAsRead(c *gin.Context) {
	userIDRaw, _ := c.Get(middleware.ContextKeyUserID)
	userID, ok := userIDRaw.(uuid.UUID)
	if !ok {
		utils.Unauthorized(c, "User not found in context")
		return
	}

	notifID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid notification ID", err)
		return
	}

	if err := h.service.MarkRead(c.Request.Context(), userID, notifID); err != nil {
		utils.InternalServerError(c, "Failed to mark notification as read", err)
		return
	}

	utils.Ok(c, "Notification marked as read", nil)
}
