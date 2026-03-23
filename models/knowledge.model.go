package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KnowledgeCategory string
type KnowledgeStatus string

const (
	CategoryGeneral      KnowledgeCategory = "general"
	CategoryInfectious   KnowledgeCategory = "infectious_diseases"
	CategoryChronic      KnowledgeCategory = "chronic_conditions"
	CategoryEmergency    KnowledgeCategory = "emergency_protocol"
	CategoryPrevention   KnowledgeCategory = "preventive_care"
	CategoryPediatrics   KnowledgeCategory = "pediatrics"

	KnowledgeStatusActive  KnowledgeStatus = "active"
	KnowledgeStatusArchived KnowledgeStatus = "archived"
)

type KnowledgeEntry struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	
	// Legacy Field Alignment (To support Seeder + Original CSVs)
	Title       string         `gorm:"column:disease_name;not null;index" json:"title"`
	Symptoms    string         `gorm:"type:text" json:"symptoms"`
	Description string         `gorm:"type:text" json:"description"`
	Causes      string         `gorm:"type:text" json:"causes"`
	Advice      string         `gorm:"type:text" json:"advice"` // Next steps
	
	// Evidence & Metadata
	Source      string         `gorm:"not null" json:"source"` // WHO, NCDC, Mayo Clinic, etc.
	Category    KnowledgeCategory `gorm:"type:varchar(50);default:'general'" json:"category"`
	Tags        string         `gorm:"type:text" json:"tags"`
	
	// Research Flags (Epidemiology)
	IsEpidemicAlert bool      `gorm:"default:false" json:"is_epidemic_alert"`
	Region          string    `gorm:"default:'Global'" json:"region"`
	UrgencyScore    int       `gorm:"default:0" json:"urgency_score"`
	IsCoreDatasset  bool      `gorm:"default:false" json:"is_core_dataset"`
	
	Status      KnowledgeStatus `gorm:"type:varchar(20);default:'active'" json:"status"`
	
	AuthorID    *uuid.UUID     `gorm:"type:uuid" json:"author_id"`
	Author      Admin          `gorm:"foreignKey:AuthorID" json:"author"`
	
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
