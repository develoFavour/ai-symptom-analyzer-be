package symptom

import (
	"log"
	"net/http"

	"ai-symptom-checker/pkg/middleware"
	"ai-symptom-checker/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler holds service dependency for the symptom module
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// getUserID is a helper to extract the authenticated user's UUID from gin context
func getUserID(c *gin.Context) (uuid.UUID, bool) {
	raw, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := raw.(uuid.UUID)
	return userID, ok
}

// StartSession POST /api/v1/symptoms/sessions
// Creates a blank session in Postgres and returns the server-generated ID
func (h *Handler) StartSession(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		utils.Unauthorized(c, "Authentication required")
		return
	}

	resp, err := h.service.CreateSession(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		utils.InternalServerError(c, "Failed to create session", err)
		return
	}

	utils.Created(c, "Session created", resp)
}

// GetSession GET /api/v1/symptoms/sessions/:sessionId
// Loads a session's persisted chat history (for page refresh restore)
func (h *Handler) GetSession(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		utils.Unauthorized(c, "Authentication required")
		return
	}

	sessionID := c.Param("sessionId")
	session, err := h.service.GetSession(c.Request.Context(), userID, sessionID)
	if err != nil {
		if err.Error() == "unauthorized" {
			utils.Forbidden(c, "You don't have access to this session")
		} else {
			utils.NotFound(c, "Session not found")
		}
		return
	}

	utils.Ok(c, "Session retrieved", session)
}

// ListSessions GET /api/v1/symptoms/sessions
// Returns all sessions belonging to the authenticated patient
func (h *Handler) ListSessions(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		utils.Unauthorized(c, "Authentication required")
		return
	}

	sessions, err := h.service.ListSessions(c.Request.Context(), userID)
	if err != nil {
		utils.InternalServerError(c, "Failed to retrieve sessions", err)
		return
	}

	utils.Ok(c, "Sessions retrieved", sessions)
}

// HandleChat POST /api/v1/symptoms/chat
// Processes one chat turn, persists to Postgres, and returns the AI response
func (h *Handler) HandleChat(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		utils.Unauthorized(c, "Authentication required")
		return
	}

	var req HandleChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.SessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	result, err := h.service.ProcessChat(c.Request.Context(), userID, req)
	if err != nil {
		log.Printf("Chat processing failed: %v", err)
		utils.InternalServerError(c, "failed to process chat prompt", err)
		return
	}

	utils.Ok(c, "Chat processed", result)
}
