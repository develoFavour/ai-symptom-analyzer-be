package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UrgencyLevel string

const (
	UrgencyEmergency UrgencyLevel = "emergency"
	UrgencySeeDoctor UrgencyLevel = "see_doctor"
	UrgencySelfCare  UrgencyLevel = "self_care"
)

type SymptomSession struct {
	ID                 uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID             uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Title              string         `gorm:"type:varchar(255)" json:"title"`
	ChatHistory        string         `gorm:"type:jsonb;default:'[]'" json:"chat_history"` // JSON array of messages
	RawInput           string         `gorm:"type:text" json:"raw_input"` // what the user originally typed
	ProcessedInput     string         `gorm:"type:text" json:"processed_input"`    // cleaned/extracted version
	UrgencyLevel       UrgencyLevel   `json:"urgency_level"`
	Status             string         `gorm:"type:varchar(50);default:'in_progress'" json:"status"` // in_progress | completed
	AIProvider         string         `json:"ai_provider"` // gemini | groq
	IsFlaggedForReview bool           `gorm:"default:false" json:"is_flagged_for_review"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User      User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Diagnoses []Diagnosis `gorm:"foreignKey:SessionID" json:"diagnoses,omitempty"`
	Feedbacks []Feedback  `gorm:"foreignKey:SessionID" json:"feedbacks,omitempty"`
}

type Diagnosis struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	SessionID     uuid.UUID `gorm:"type:uuid;not null;index" json:"session_id"`
	ConditionName string    `gorm:"not null" json:"condition_name"`
	Description   string    `gorm:"type:text" json:"description"`
	Confidence    string    `json:"confidence"` // high | medium | low
	CommonCauses  string    `gorm:"type:text" json:"common_causes"`
	HealthAdvice  string    `gorm:"type:text" json:"health_advice"`
	Rank          int       `json:"rank"` // 1 = top match
	CreatedAt     time.Time `json:"created_at"`
}
