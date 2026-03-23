package admin

import (
	"net/http"
	"strconv"

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

// ── Dashboard & Stats ────────────────────────────────────────────────────────

func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve stats", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "System stats retrieved", stats)
}

// ── Doctor Management ────────────────────────────────────────────────────────

func (h *Handler) InviteDoctor(c *gin.Context) {
	var req InviteDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	adminIDRaw, _ := c.Get(middleware.ContextKeyUserID)
	adminID, ok := adminIDRaw.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context missing", nil)
		return
	}

	if err := h.service.InviteDoctor(c.Request.Context(), adminID, req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, "Doctor invited successfully", nil)
}

func (h *Handler) ListDoctors(c *gin.Context) {
	doctors, err := h.service.ListDoctors(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve doctors", err)
		return
	}
	utils.Ok(c, "Doctors retrieved", doctors)
}

func (h *Handler) UpdateDoctorStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminIDRaw, _ := c.Get(middleware.ContextKeyUserID)
	adminID := adminIDRaw.(uuid.UUID)

	if err := h.service.UpdateDoctorStatus(c.Request.Context(), adminID, id, req.Status); err != nil {
		utils.InternalServerError(c, "Failed to update doctor status", err)
		return
	}
	utils.Ok(c, "Doctor status updated", nil)
}

// ── User Management ──────────────────────────────────────────────────────────

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.service.ListUsers(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve users", err)
		return
	}
	utils.Ok(c, "Users retrieved", users)
}

func (h *Handler) UpdateUserStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		IsActive string `json:"is_active" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isActive, _ := strconv.ParseBool(req.IsActive)
	adminIDRaw, _ := c.Get(middleware.ContextKeyUserID)
	adminID := adminIDRaw.(uuid.UUID)

	if err := h.service.UpdateUserStatus(c.Request.Context(), adminID, id, isActive); err != nil {
		utils.InternalServerError(c, "Failed to update user status", err)
		return
	}
	utils.Ok(c, "User status updated", nil)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	adminIDRaw, _ := c.Get(middleware.ContextKeyUserID)
	adminID := adminIDRaw.(uuid.UUID)

	if err := h.service.DeleteUser(c.Request.Context(), adminID, id); err != nil {
		utils.InternalServerError(c, "Failed to remove user account", err)
		return
	}
	utils.Ok(c, "User profile permanently deleted", nil)
}

// ── Admin Onboarding ─────────────────────────────────────────────────────────

func (h *Handler) InviteAdmin(c *gin.Context) {
	var req InviteAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	adminIDRaw, _ := c.Get(middleware.ContextKeyUserID)
	adminID := adminIDRaw.(uuid.UUID)

	if err := h.service.InviteAdmin(c.Request.Context(), adminID, req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, "Administrator invitation sent", nil)
}

func (h *Handler) ListAdmins(c *gin.Context) {
	admins, err := h.service.ListAdmins(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve administrators", err)
		return
	}
	utils.Ok(c, "Administrators retrieved", admins)
}

func (h *Handler) DeleteAdmin(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admin id"})
		return
	}

	actorIDRaw, _ := c.Get(middleware.ContextKeyUserID)
	actorID := actorIDRaw.(uuid.UUID)

	if err := h.service.DeleteAdmin(c.Request.Context(), actorID, id); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	utils.Ok(c, "Administrator removed", nil)
}
