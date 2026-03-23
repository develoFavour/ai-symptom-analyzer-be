package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DoctorStatus string

const (
	DoctorStatusPending   DoctorStatus = "pending" // invite sent, not yet set up
	DoctorStatusActive    DoctorStatus = "active"
	DoctorStatusSuspended DoctorStatus = "suspended"
)

type Doctor struct {
	ID               uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name             string         `gorm:"not null" json:"name"`
	Email            string         `gorm:"uniqueIndex;not null" json:"email"`
	Password         string         `gorm:"not null" json:"-"`
	Specialization   string         `json:"specialization"`
	Credentials      string         `json:"credentials"`
	Bio              string         `gorm:"type:text" json:"bio"`
	SetupComplete    bool           `gorm:"default:false" json:"setup_complete"`
	InviteToken      string         `gorm:"index" json:"-"` // used for account setup link
	InviteExpiry     *time.Time     `json:"-"`
	Status           DoctorStatus   `gorm:"default:'pending'" json:"status"`
	IsActive         bool           `gorm:"default:true" json:"is_active"`
	InvitedByAdminID uuid.UUID      `json:"invited_by_admin_id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Consultations []Consultation `gorm:"foreignKey:DoctorID" json:"consultations,omitempty"`
	Notifications []Notification `gorm:"foreignKey:DoctorID" json:"notifications,omitempty"`
}
