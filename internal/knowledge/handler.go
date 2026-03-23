package knowledge

import (
	"ai-symptom-checker/models"
	"ai-symptom-checker/pkg/middleware"
	"ai-symptom-checker/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateEntry(c *gin.Context) {
	var req CreateEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid entry format", err)
		return
	}

	adminID, _ := c.Get(middleware.ContextKeyUserID)
	entry, err := h.service.CreateEntry(c.Request.Context(), adminID.(uuid.UUID), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create knowledge entry", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Medical entry recorded", entry)
}

func (h *Handler) GetEntries(c *gin.Context) {
	params := utils.GetPagination(c)
	filter := make(map[string]interface{})
	if cat := c.Query("category"); cat != "" {
		filter["category"] = cat
	}
	if alert := c.Query("is_epidemic_alert"); alert != "" {
		filter["is_epidemic_alert"] = alert == "true"
	}
	if search := c.Query("search"); search != "" {
		filter["search"] = search
	}

	role, _ := c.Get(middleware.ContextKeyRole)

	var (
		entries []models.KnowledgeEntry
		total   int64
		err     error
	)

	if role == "doctor" {
		entries, total, err = h.service.InquireEntries(c.Request.Context(), filter, params.Limit, params.Offset)
	} else {
		filter["is_active"] = true
		entries, total, err = h.service.ListEntries(c.Request.Context(), filter, params.Limit, params.Offset)
	}

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Inquiry failed", err)
		return
	}

	utils.Ok(c, "Verified directory retrieved", utils.BuildPaginatedResponse(entries, total, params))
}

func (h *Handler) GetAlerts(c *gin.Context) {
	alerts, err := h.service.GetGlobalThreats(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch threat telemetry", err)
		return
	}
	utils.Ok(c, "Active epidemiological threats", alerts)
}

func (h *Handler) UpdateEntry(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid identity token", nil)
		return
	}

	var req CreateEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid payload", err)
		return
	}

	if err := h.service.UpdateEntry(c.Request.Context(), id, req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Correction failed", err)
		return
	}

	utils.Ok(c, "Medical entry revised", nil)
}

func (h *Handler) DeleteEntry(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid identity token", nil)
		return
	}

	if err := h.service.DeleteEntry(c.Request.Context(), id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Deactivation failed", err)
		return
	}

	utils.Ok(c, "Medical entry decommissioned", nil)
}
