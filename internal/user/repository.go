package user

import (
    "ai-symptom-checker/models"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

// Repository defines the DB contract for the user module
type Repository interface {
	GetDashboardStats(userID uuid.UUID) (map[string]interface{}, error)
	GetRecentHistory(userID uuid.UUID, limit int) ([]interface{}, error)
	GetActiveConsultation(userID uuid.UUID) (*models.Consultation, error)
}

type postgresRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) GetDashboardStats(userID uuid.UUID) (map[string]interface{}, error) {
	var symptomCount int64
	var consultCount int64

	r.db.Model(&models.SymptomSession{}).Where("user_id = ?", userID).Count(&symptomCount)
	r.db.Model(&models.Consultation{}).Where("user_id = ?", userID).Count(&consultCount)

	// Calculate a dummy health score or based on data
	healthScore := "Good"
	if symptomCount > 5 {
		healthScore = "Fair"
	}

	return map[string]interface{}{
		"symptom_checks": symptomCount,
		"consultations":  consultCount,
		"health_score":   healthScore,
	}, nil
}

func (r *postgresRepository) GetRecentHistory(userID uuid.UUID, limit int) ([]interface{}, error) {
    var sessions []models.SymptomSession
    err := r.db.Where("user_id = ?", userID).
        Order("created_at DESC").
        Limit(limit).
        Find(&sessions).Error
    if err != nil {
        return nil, err
    }

    var history []interface{}
    for _, s := range sessions {
        history = append(history, map[string]interface{}{
            "id":         s.ID,
            "type":       "symptom_check",
            "title":      s.Title,
            "date":       s.CreatedAt,
            "urgency":    s.UrgencyLevel,
            "status":     s.Status,
        })
    }

    return history, nil
}

func (r *postgresRepository) GetActiveConsultation(userID uuid.UUID) (*models.Consultation, error) {
	var consult models.Consultation
	err := r.db.Where("user_id = ? AND status IN ?", userID, []string{"open", "responded"}).
		Order("created_at DESC").
		First(&consult).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &consult, nil
}
