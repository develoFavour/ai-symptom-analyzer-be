package symptom

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"ai-symptom-checker/models"
	"ai-symptom-checker/pkg/ai"

	"github.com/google/uuid"
)

// HandleChatRequest payload structure from frontend
type HandleChatRequest struct {
	SessionID      string           `json:"session_id"`
	PatientContext string           `json:"patient_context"`
	ChatHistory    []ai.ChatMessage `json:"chat_history"`
}

// CreateSessionResponse is returned to the frontend on session init
type CreateSessionResponse struct {
	SessionID string `json:"session_id"`
}

type Service struct {
	repo     Repository
	aiClient ai.Client
}

func NewService(repo Repository, aiClient ai.Client) *Service {
	return &Service{
		repo:     repo,
		aiClient: aiClient,
	}
}

// CreateSession initializes a blank session on the server and returns its ID
func (s *Service) CreateSession(ctx context.Context, userID uuid.UUID) (*CreateSessionResponse, error) {
	session := &models.SymptomSession{
		UserID: userID,
		Title:  "New Conversation",
		Status: "in_progress",
	}
	if err := s.repo.InitSession(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return &CreateSessionResponse{SessionID: session.ID.String()}, nil
}

// GetSession returns a session record (used to load history on page refresh)
func (s *Service) GetSession(ctx context.Context, userID uuid.UUID, sessionIDStr string) (*models.SymptomSession, error) {
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid session id")
	}
	session, err := s.repo.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}
	// Ensure user can only read their own sessions
	if session.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}
	return session, nil
}

// ListSessions returns all sessions for the authenticated user
func (s *Service) ListSessions(ctx context.Context, userID uuid.UUID) ([]models.SymptomSession, error) {
	return s.repo.ListUserSessions(userID)
}

// ProcessChat handles one chat turn for a given session
func (s *Service) ProcessChat(ctx context.Context, userID uuid.UUID, req HandleChatRequest) (*ai.AnalysisResult, error) {
	// Parse Session ID
	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session_id")
	}

	// Security check: session must belong to this user
	session, err := s.repo.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}
	if session.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	// Extract all user messages to build a comprehensive clinical query
	userConcatenatedQuery := ""
	for _, msg := range req.ChatHistory {
		if msg.Role == "user" {
			userConcatenatedQuery += " " + msg.Content
		}
	}
	if userConcatenatedQuery == "" {
		return nil, fmt.Errorf("no user messages found")
	}
	userConcatenatedQuery = strings.TrimSpace(userConcatenatedQuery)

	// 1. Search the medical knowledge base (RAG) with full history context
	entries, err := s.repo.SearchKnowledgeBase(userConcatenatedQuery)
	if err != nil {
		log.Printf("KnowledgeBase search error: %v", err)
	}

	// 2. Format DB entries to string for RAG context
	var knowledgeData string
	for _, entry := range entries {
		knowledgeData += fmt.Sprintf("Source: %s\nCondition: %s\nSymptoms: %s\nAnalysis: %s\nAdvice: %s\n\n",
			entry.Source, entry.Title, entry.Symptoms, entry.Description, entry.Advice)
	}
	if knowledgeData == "" {
		knowledgeData = "No matching records found. Ask clarifying questions and respond safely."
	}

	// 3. Call the LLM
	aiReq := ai.AnalysisRequest{
		PatientContext: req.PatientContext,
		ChatHistory:    req.ChatHistory,
		KnowledgeData:  knowledgeData,
	}
	result, err := s.aiClient.AnalyzeChat(ctx, aiReq)
	if err != nil {
		return nil, fmt.Errorf("AI analyze error: %w", err)
	}

	// 4. Build the updated history (incoming + AI reply) and persist it
	updatedHistory := append(req.ChatHistory, ai.ChatMessage{
		Role:    "model",
		Content: result.BotMessage,
	})
	historyJSON, _ := json.Marshal(updatedHistory)

	// Auto-generate a readable title from the first user message (if still default)
	title := session.Title
	if title == "" || title == "New Conversation" {
		firstUserMsg := userConcatenatedQuery
		if len(firstUserMsg) > 55 {
			firstUserMsg = firstUserMsg[:55] + "..."
		}
		title = firstUserMsg
	}

	// Persist chat history + title update
	if err := s.repo.UpdateSessionChat(sessionID, string(historyJSON), title); err != nil {
		log.Printf("Failed to update session chat: %v", err)
	}

	// 5. If diagnosis ready, finalize the session with urgency and diagnoses
	if result.IsDiagnosisReady {
		log.Printf("[Session %s] Diagnosis ready — finalizing", sessionID)

		urgency := models.UrgencySelfCare
		if result.UrgencyLevel == "emergency" {
			urgency = models.UrgencyEmergency
		} else if result.UrgencyLevel == "see_doctor" {
			urgency = models.UrgencySeeDoctor
		}

		var diagnoses []models.Diagnosis
		for i, cond := range result.PossibleConditions {
			diagnoses = append(diagnoses, models.Diagnosis{
				SessionID:     sessionID,
				ConditionName: cond.Name,
				Description:   cond.Description,
				Confidence:    cond.Confidence,
				CommonCauses:  strings.Join(cond.CommonCauses, ", "),
				HealthAdvice:  result.HealthAdvice,
				Rank:          i + 1,
			})
		}

		if err := s.repo.FinalizeSession(sessionID, urgency, diagnoses); err != nil {
			log.Printf("Failed to finalize session: %v", err)
		}
	}

	return result, nil
}
