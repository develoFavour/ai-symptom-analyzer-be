package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConsultationStatus string
type ConsultationUrgency string
type SharingMode string

const (
	ConsultationStatusPending  ConsultationStatus = "pending"
	ConsultationStatusAnswered ConsultationStatus = "answered"
	ConsultationStatusEscalated ConsultationStatus = "escalated"
	ConsultationStatusClosed   ConsultationStatus = "closed"

	ConsultationUrgencyRoutine ConsultationUrgency = "routine"
	ConsultationUrgencySoon    ConsultationUrgency = "soon"
	ConsultationUrgencyUrgent  ConsultationUrgency = "urgent"

	SharingModeFull    SharingMode = "full"    // Doctor can read the full AI chat transcript
	SharingModeSummary SharingMode = "summary" // Doctor only sees AI-generated clinical summary
)

type Consultation struct {
	ID             uuid.UUID           `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID         uuid.UUID           `gorm:"type:uuid;not null;index" json:"user_id"`
	DoctorID       *uuid.UUID          `gorm:"type:uuid;index" json:"doctor_id"`           // chosen by patient upfront
	SessionID      *uuid.UUID          `gorm:"type:uuid;index" json:"session_id"`           // linked AI chat session
	Symptoms       string              `gorm:"type:text;not null" json:"symptoms"`          // pre-filled symptom summary
	PatientNote    string              `gorm:"type:text" json:"patient_note"`               // patient's personal message to doctor
	Urgency        ConsultationUrgency `gorm:"default:'routine'" json:"urgency"`
	Status         ConsultationStatus  `gorm:"default:'pending'" json:"status"`

	// Privacy / Data Sharing
	SharingMode    SharingMode `gorm:"type:varchar(20);default:'summary'" json:"sharing_mode"` // "full" | "summary"
	SharedSummary  string      `gorm:"type:text" json:"shared_summary"`                        // AI-generated clinical summary (for "summary" mode)

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User     User                  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Doctor   *Doctor               `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	Session  *SymptomSession       `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	Replies  []ConsultationReply   `gorm:"foreignKey:ConsultationID" json:"replies,omitempty"`
	Messages []ConsultationMessage `gorm:"foreignKey:ConsultationID" json:"messages,omitempty"`
}

type ConsultationReply struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ConsultationID uuid.UUID `gorm:"type:uuid;not null;index" json:"consultation_id"`
	DoctorID       uuid.UUID `gorm:"type:uuid;not null" json:"doctor_id"`
	ReplyText      string    `gorm:"type:text;not null" json:"reply_text"`
	Recommendation string    `gorm:"type:text" json:"recommendation"` // e.g. "See a neurologist"
	IsAICorrection bool      `gorm:"default:false" json:"is_ai_correction"`
	CreatedAt      time.Time `json:"created_at"`

	// Relations
	Doctor Doctor `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
}

type ConsultationMessage struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ConsultationID uuid.UUID `gorm:"type:uuid;not null;index" json:"consultation_id"`
	SenderID       uuid.UUID `gorm:"type:uuid;not null" json:"sender_id"`
	SenderRole     string    `gorm:"type:varchar(20);not null" json:"sender_role"` // "doctor" | "patient"
	Content        string    `gorm:"type:text;not null" json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}
