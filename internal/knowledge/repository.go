package knowledge

import (
	"context"
	"ai-symptom-checker/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateEntry(ctx context.Context, entry *models.KnowledgeEntry) error
	GetEntryByID(ctx context.Context, id uuid.UUID) (*models.KnowledgeEntry, error)
	ListEntries(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]models.KnowledgeEntry, int64, error)
	UpdateEntry(ctx context.Context, entry *models.KnowledgeEntry) error
	DeleteEntry(ctx context.Context, id uuid.UUID) error
	
	// Open Enquiry for Doctors
	InquireEntries(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]models.KnowledgeEntry, int64, error)
	
	// Intelligent Search for RAG
	SearchSimilar(ctx context.Context, query string, limit int) ([]models.KnowledgeEntry, error)
	GetActiveEpidemicAlerts(ctx context.Context) ([]models.KnowledgeEntry, error)
}

type postgresRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateEntry(ctx context.Context, entry *models.KnowledgeEntry) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *postgresRepository) GetEntryByID(ctx context.Context, id uuid.UUID) (*models.KnowledgeEntry, error) {
	var entry models.KnowledgeEntry
	err := r.db.WithContext(ctx).First(&entry, id).Error
	return &entry, err
}

func (r *postgresRepository) ListEntries(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]models.KnowledgeEntry, int64, error) {
	var entries []models.KnowledgeEntry
	var total int64
	query := r.db.WithContext(ctx).Model(&models.KnowledgeEntry{}).Preload("Author")
	
	if cat, ok := filter["category"]; ok {
		query = query.Where("category = ?", cat)
	}
	if status, ok := filter["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if alert, ok := filter["is_epidemic_alert"]; ok {
		query = query.Where("is_epidemic_alert = ?", alert)
	}
	if search, ok := filter["search"]; ok && search != "" {
		s := "%" + search.(string) + "%"
		query = query.Where("disease_name ILIKE ? OR description ILIKE ?", s, s)
	}
	
	// Count total matches before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("disease_name asc").Limit(limit).Offset(offset).Find(&entries).Error
	return entries, total, err
}

func (r *postgresRepository) InquireEntries(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]models.KnowledgeEntry, int64, error) {
	return r.ListEntries(ctx, filter, limit, offset)
}

func (r *postgresRepository) UpdateEntry(ctx context.Context, entry *models.KnowledgeEntry) error {
	return r.db.WithContext(ctx).Save(entry).Error
}

func (r *postgresRepository) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.KnowledgeEntry{}, id).Error
}

func (r *postgresRepository) SearchSimilar(ctx context.Context, query string, limit int) ([]models.KnowledgeEntry, error) {
	var entries []models.KnowledgeEntry
	// Simple ILIKE search for title, content and tags for now.
	// In the future this could be upgraded to full-text or vector search.
	err := r.db.WithContext(ctx).
		Where("status = ?", models.KnowledgeStatusActive).
		Where("disease_name ILIKE ? OR description ILIKE ? OR tags ILIKE ?", 
			"%"+query+"%", "%"+query+"%", "%"+query+"%").
		Limit(limit).
		Find(&entries).Error
	return entries, err
}

func (r *postgresRepository) GetActiveEpidemicAlerts(ctx context.Context) ([]models.KnowledgeEntry, error) {
	var alerts []models.KnowledgeEntry
	err := r.db.WithContext(ctx).
		Where("status = ? AND is_epidemic_alert = ?", models.KnowledgeStatusActive, true).
		Order("urgency_score desc, created_at desc").
		Find(&alerts).Error
	return alerts, err
}
