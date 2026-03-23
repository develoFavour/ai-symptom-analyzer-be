package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type AdminStatus string

const (
    AdminStatusActive  AdminStatus = "active"
    AdminStatusPending AdminStatus = "pending"
    AdminStatusSuspended AdminStatus = "suspended"
)

type Admin struct {
    ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
    Name         string         `gorm:"not null" json:"name"`
    Username     string         `gorm:"uniqueIndex" json:"username"` // Allowed to be empty during invite
    Email        string         `gorm:"uniqueIndex;not null" json:"email"`
    Password     string         `gorm:"not null" json:"-"`
    Status       AdminStatus    `gorm:"type:varchar(20);default:'active'" json:"status"`
    IsActive     bool           `gorm:"default:true" json:"is_active"`
    InviteToken  string         `gorm:"uniqueIndex" json:"invite_token"`
    InviteExpiry *time.Time     `json:"invite_expiry"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
