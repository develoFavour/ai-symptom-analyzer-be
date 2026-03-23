package consultation

import (
	"context"
	"encoding/json"
	"fmt"

	"ai-symptom-checker/internal/notification"
	"ai-symptom-checker/models"
	"ai-symptom-checker/pkg/ai"
	"github.com/google/uuid"
)

// CreateConsultationRequest is the payload from the patient
// CreateConsultationRequest is the payload from the patient
type CreateConsultationRequest struct {
	DoctorID    string `json:"doctor_id" binding:"required"`
	SessionID   string `json:"session_id"`   // the AI chat session this is about
	Symptoms    string `json:"symptoms" binding:"required"`
	PatientNote string `json:"patient_note"`
	Urgency     string `json:"urgency"`      // routine | soon | urgent
	SharingMode string `json:"sharing_mode"` // full | summary
}

// DoctorReplyRequest is the payload when a doctor responds
type DoctorReplyRequest struct {
	ReplyText      string `json:"reply_text" binding:"required"`
	Recommendation string `json:"recommendation"`
	IsAICorrection bool   `json:"is_ai_correction"`
}

type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

type Service struct {
	repo     Repository
	aiClient ai.Client
	notifSvc *notification.Service
}

func NewService(repo Repository, aiClient ai.Client, notifSvc *notification.Service) *Service {
	return &Service{repo: repo, aiClient: aiClient, notifSvc: notifSvc}
}

// ListDoctors returns all active doctors for the patient to browse
func (s *Service) ListDoctors(ctx context.Context) ([]models.Doctor, error) {
	return s.repo.ListActiveDoctors()
}

// CreateConsultation creates the consultation request, optionally generating a clinical summary
func (s *Service) CreateConsultation(ctx context.Context, userID uuid.UUID, req CreateConsultationRequest) (*models.Consultation, error) {
	doctorID, err := uuid.Parse(req.DoctorID)
	if err != nil {
		return nil, fmt.Errorf("invalid doctor_id")
	}

	// Default values
	urgency := models.ConsultationUrgencyRoutine
	switch req.Urgency {
	case "soon":
		urgency = models.ConsultationUrgencySoon
	case "urgent":
		urgency = models.ConsultationUrgencyUrgent
	}

	sharingMode := models.SharingModeSummary
	if req.SharingMode == "full" {
		sharingMode = models.SharingModeFull
	}

	consultation := &models.Consultation{
		UserID:      userID,
		DoctorID:    &doctorID,
		Symptoms:    req.Symptoms,
		PatientNote: req.PatientNote,
		Urgency:     urgency,
		SharingMode: sharingMode,
	}

	// Link to AI session if provided
	if req.SessionID != "" {
		sessionID, err := uuid.Parse(req.SessionID)
		if err == nil {
			consultation.SessionID = &sessionID
		}
	}

	// If sharing mode is "summary", generate and store a clinical AI summary now
	if sharingMode == models.SharingModeSummary && consultation.SessionID != nil {
		chatHistoryJSON, err := s.repo.GetSessionChatHistory(*consultation.SessionID)
		if err == nil && chatHistoryJSON != "" && chatHistoryJSON != "[]" {
			var history []ai.ChatMessage
			if jsonErr := json.Unmarshal([]byte(chatHistoryJSON), &history); jsonErr == nil {
				sumReq := ai.SummarizeRequest{
					ChatHistory:    history,
					PatientContext: fmt.Sprintf("Patient note: %s", req.PatientNote),
				}
				summary, sumErr := s.aiClient.GenerateClinicalSummary(ctx, sumReq)
				if sumErr == nil {
					consultation.SharedSummary = summary
				}
			}
		}
	}

	if err := s.repo.Create(consultation); err != nil {
		return nil, fmt.Errorf("failed to create consultation: %w", err)
	}

	return consultation, nil
}

// GetMyConsultations returns all the patient's previous consultation requests
func (s *Service) GetMyConsultations(ctx context.Context, userID uuid.UUID) ([]models.Consultation, error) {
	return s.repo.ListByPatient(userID)
}

// GetConsultationDetails gets a single consultation, ensuring authorization
func (s *Service) GetConsultationDetails(ctx context.Context, id uuid.UUID, actorID uuid.UUID, isDoctor bool) (*models.Consultation, error) {
	c, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Authorization check: Must be the patient or the assigned doctor
	if !isDoctor && c.UserID != actorID {
		return nil, fmt.Errorf("unauthorized")
	}
	if isDoctor && (c.DoctorID == nil || *c.DoctorID != actorID) {
		return nil, fmt.Errorf("unauthorized")
	}

	return c, nil
}

// ListDoctorQueue returns all consultations assigned to a specific doctor
func (s *Service) ListDoctorQueue(ctx context.Context, doctorID uuid.UUID) ([]models.Consultation, error) {
	return s.repo.ListByDoctor(doctorID)
}

// ReplyToConsultation adds a doctor's response and updates the status
func (s *Service) ReplyToConsultation(ctx context.Context, doctorID uuid.UUID, consultationID uuid.UUID, req DoctorReplyRequest) (*models.ConsultationReply, error) {
	// 1. Verify consultation belongs to this doctor
	c, err := s.repo.GetByID(consultationID)
	if err != nil {
		return nil, fmt.Errorf("consultation not found")
	}

	if c.DoctorID == nil || *c.DoctorID != doctorID {
		return nil, fmt.Errorf("not assigned to you")
	}

	// 2. Create the reply
	reply := &models.ConsultationReply{
		ConsultationID: consultationID,
		DoctorID:       doctorID,
		ReplyText:      req.ReplyText,
		Recommendation: req.Recommendation,
		IsAICorrection: req.IsAICorrection,
	}

	if err := s.repo.CreateReply(reply); err != nil {
		return nil, err
	}

	// 3. Mark consultation as answered
	if err := s.repo.UpdateStatus(consultationID, models.ConsultationStatusAnswered); err != nil {
		return nil, err
	}

	// 4. Send notification to patient
	if s.notifSvc != nil {
		s.notifSvc.CreateNotification(ctx, &models.Notification{
			UserID:  &c.UserID,
			Type:    models.NotifConsultationAnswered,
			Title:   "Consultation Answered",
			Message: fmt.Sprintf("Dr. %s has responded to your consultation request.", c.Doctor.Name),
			RefID:   &consultationID,
		})
	}

	return reply, nil
}

// AddMessage adds a simple chat message (from either patient or doctor)
func (s *Service) AddMessage(ctx context.Context, actorID uuid.UUID, role string, consultationID uuid.UUID, req SendMessageRequest) (*models.ConsultationMessage, error) {
	// 1. Authorization: verify consultation belongs to this actor (as either user or doctor)
	c, err := s.repo.GetByID(consultationID)
	if err != nil {
		return nil, fmt.Errorf("consultation not found")
	}

	isAuthorized := false
	if role == "doctor" && c.DoctorID != nil && *c.DoctorID == actorID {
		isAuthorized = true
	} else if role == "patient" && c.UserID == actorID {
		isAuthorized = true
	}

	if !isAuthorized {
		return nil, fmt.Errorf("unauthorized to message this consultation")
	}

	// 2. Create message
	msg := &models.ConsultationMessage{
		ConsultationID: consultationID,
		SenderID:       actorID,
		SenderRole:     role,
		Content:        req.Content,
	}

	if err := s.repo.CreateMessage(msg); err != nil {
		return nil, err
	}

	// 3. Optional: Add a notification for the other party
	if s.notifSvc != nil {
		targetUserID := c.UserID
		title := "New Message from Doctor"
		if role == "patient" {
			if c.DoctorID != nil {
				targetUserID = *c.DoctorID
				title = "New Message from Patient"
			} else {
				// No doctor assigned yet, skip notification for now
				return msg, nil
			}
		}

		s.notifSvc.CreateNotification(ctx, &models.Notification{
			UserID:  &targetUserID,
			Type:    models.NotifConsultationAnswered, // reusing for now
			Title:   title,
			Message: "A new message was added to your consultation.",
			RefID:   &consultationID,
		})
	}

	return msg, nil
}
