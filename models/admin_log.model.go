package models

import (
    "time"
    "github.com/google/uuid"
)

type AdminLog struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	AdminID   uuid.UUID `gorm:"type:uuid;not null;index" json:"admin_id"`
	Action    string    `gorm:"not null" json:"action"`     // invited_doctor, suspended_doctor, updated_knowledge, etc.
	TargetID  string    `json:"target_id,omitempty"`       // ID of the target resource (doctor_id, etc.)
	Details   string    `json:"details"`                   // "Invited Dr. Evans", "Suspended user@test.com"
	CreatedAt time.Time `json:"created_at"`

	// Relations
	Admin Admin `gorm:"foreignKey:AdminID" json:"admin,omitempty"`
}
