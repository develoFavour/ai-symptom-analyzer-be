package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID               uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name             string         `gorm:"not null" json:"name"`
	Email            string         `gorm:"uniqueIndex;not null" json:"email"`
	Password         string         `gorm:"not null" json:"-"`
	Age              int            `json:"age"`
	Gender           string         `gorm:"type:varchar(10)" json:"gender"` // male | female | other
	KnownAllergies   string         `json:"known_allergies,omitempty"`
	PreExistingConds string         `json:"pre_existing_conditions,omitempty"`
	IsEmailVerified    bool           `gorm:"default:false" json:"is_email_verified"`
	VerificationToken string         `gorm:"index" json:"-"`
	VerificationExpiry *time.Time     `json:"-"`
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Sessions      []SymptomSession `gorm:"foreignKey:UserID" json:"sessions,omitempty"`
	Consultations []Consultation   `gorm:"foreignKey:UserID" json:"consultations,omitempty"`
	Notifications []Notification   `gorm:"foreignKey:UserID" json:"notifications,omitempty"`
}
