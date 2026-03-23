package consultation

import (
	"net/http"

	"ai-symptom-checker/pkg/middleware"
	"ai-symptom-checker/pkg/socket"
	"ai-symptom-checker/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
	hub     *socket.Hub
}

func NewHandler(service *Service, hub *socket.Hub) *Handler {
	return &Handler{service: service, hub: hub}
}

func getUserID(c *gin.Context) (uuid.UUID, bool) {
	raw, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := raw.(uuid.UUID)
	return userID, ok
}

func getRole(c *gin.Context) string {
	raw, exists := c.Get(middleware.ContextKeyRole)
	if !exists {
		return ""
	}
	role, _ := raw.(string)
	return role
}

// ListDoctors GET /api/v1/consultations/doctors
// Returns all active doctors available for consultation
func (h *Handler) ListDoctors(c *gin.Context) {
	doctors, err := h.service.ListDoctors(c.Request.Context())
	if err != nil {
		utils.InternalServerError(c, "Failed to fetch doctors", err)
		return
	}
	utils.Ok(c, "Doctors retrieved", doctors)
}

// CreateConsultation POST /api/v1/consultations
// Patient submits a consultation request to a specific doctor
func (h *Handler) CreateConsultation(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		utils.Unauthorized(c, "Authentication required")
		return
	}

	var req CreateConsultationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	consultation, err := h.service.CreateConsultation(c.Request.Context(), userID, req)
	if err != nil {
		utils.InternalServerError(c, err.Error(), err)
		return
	}

	utils.Created(c, "Consultation request submitted successfully", consultation)
}

// GetMyConsultations GET /api/v1/consultations
// Returns all consultation requests made by the authenticated patient
func (h *Handler) GetMyConsultations(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		utils.Unauthorized(c, "Authentication required")
		return
	}

	list, err := h.service.GetMyConsultations(c.Request.Context(), userID)
	if err != nil {
		utils.InternalServerError(c, "Failed to fetch consultations", err)
		return
	}

	utils.Ok(c, "Consultations retrieved", list)
}

// GetConsultation GET /api/v1/consultations/:id
// Returns details for a specific consultation
func (h *Handler) GetConsultation(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		utils.Unauthorized(c, "Authentication required")
		return
	}

	role := getRole(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid consultation id"})
		return
	}

	consult, err := h.service.GetConsultationDetails(c.Request.Context(), id, userID, role == "doctor")
	if err != nil {
		if err.Error() == "unauthorized" {
			utils.Forbidden(c, "You do not have permission to view this consultation")
		} else {
			utils.InternalServerError(c, "Failed to fetch consultation details", err)
		}
		return
	}

	utils.Ok(c, "Consultation details retrieved", consult)
}

// ListDoctorQueue GET /api/v1/doctor/consultations
// Returns all consultations assigned to the logged-in doctor
func (h *Handler) ListDoctorQueue(c *gin.Context) {
	doctorID, ok := getUserID(c)
	if !ok {
		utils.Unauthorized(c, "Doctor authentication required")
		return
	}

	queue, err := h.service.ListDoctorQueue(c.Request.Context(), doctorID)
	if err != nil {
		utils.InternalServerError(c, "Failed to fetch doctor queue", err)
		return
	}

	utils.Ok(c, "Doctor queue retrieved", queue)
}

// ReplyToConsultation POST /api/v1/doctor/consultations/:id/reply
// Doctor submits a reply to a consultation
func (h *Handler) ReplyToConsultation(c *gin.Context) {
	doctorID, ok := getUserID(c)
	if !ok {
		utils.Unauthorized(c, "Doctor authentication required")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid consultation id"})
		return
	}

	var req DoctorReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reply, err := h.service.ReplyToConsultation(c.Request.Context(), doctorID, id, req)
	if err != nil {
		utils.InternalServerError(c, err.Error(), err)
		return
	}

	utils.Ok(c, "Reply submitted successfully", reply)
}

// AddMessage POST /api/v1/consultations/:id/messages
// Send chat messages back and forth in a consultation
func (h *Handler) AddMessage(c *gin.Context) {
	actorID, ok := getUserID(c)
	if !ok {
		utils.Unauthorized(c, "Authentication required")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid consultation id"})
		return
	}

	role := getRole(c)

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, err := h.service.AddMessage(c.Request.Context(), actorID, role, id, req)
	if err != nil {
		utils.InternalServerError(c, err.Error(), err)
		return
	}

	// Real-time broadcast
	h.hub.BroadcastToRoom(id, "new_message", msg)

	utils.Ok(c, "Message sent", msg)
}

// WebSocketUpgrade handles the WS connection
func (h *Handler) WebSocketUpgrade(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return
	}

	// Optional: Verifiy user has access to this consultation (auth middleware already ran)
	h.hub.HandleWebSocket(c.Writer, c.Request, id)
}
