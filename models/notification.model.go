package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotifConsultationAnswered NotificationType = "consultation_answered"
	NotifCaseReviewed         NotificationType = "case_reviewed"
	NotifDoctorApproved       NotificationType = "doctor_approved"
	NotifSystemAlert          NotificationType = "system_alert"
)

type Notification struct {
	ID        uuid.UUID        `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    *uuid.UUID       `gorm:"type:uuid;index" json:"user_id,omitempty"`   // patient
	DoctorID  *uuid.UUID       `gorm:"type:uuid;index" json:"doctor_id,omitempty"` // doctor
	Type      NotificationType `gorm:"not null" json:"type"`
	Title     string           `gorm:"not null" json:"title"`
	Message   string           `gorm:"type:text;not null" json:"message"`
	IsRead    bool             `gorm:"default:false" json:"is_read"`
	RefID     *uuid.UUID       `gorm:"type:uuid" json:"ref_id,omitempty"` // e.g. consultation ID
	CreatedAt time.Time        `json:"created_at"`
}

type Feedback struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	SessionID uuid.UUID `gorm:"type:uuid;not null;index" json:"session_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Helpful   bool      `json:"helpful"`
	Note      string    `gorm:"type:text" json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
