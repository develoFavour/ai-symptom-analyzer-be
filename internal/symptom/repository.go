package symptom

import (
	"ai-symptom-checker/models"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines the DB contract for the symptom module
type Repository interface {
	SearchKnowledgeBase(query string) ([]models.KnowledgeEntry, error)

	// Session management
	InitSession(session *models.SymptomSession) error
	GetSession(sessionID uuid.UUID) (*models.SymptomSession, error)
	UpdateSessionChat(sessionID uuid.UUID, chatHistoryJSON string, title string) error
	FinalizeSession(sessionID uuid.UUID, urgency models.UrgencyLevel, diagnoses []models.Diagnosis) error
	ListUserSessions(userID uuid.UUID) ([]models.SymptomSession, error)
}

type postgresRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) SearchKnowledgeBase(query string) ([]models.KnowledgeEntry, error) {
	var entries []models.KnowledgeEntry

	// Split query into words to build an OR-based high-recall query
	words := strings.Fields(query)
	var tsQueryParts []string
	for _, word := range words {
		// Only use words longer than 2 characters (ignoring "a", "is", "I", etc.)
		if len(word) > 2 {
			// Sanitize for tsquery (remove any non-alphanumeric)
			clean := strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
					return r
				}
				return -1
			}, word)
			if len(clean) > 2 {
				tsQueryParts = append(tsQueryParts, fmt.Sprintf("%s:*", clean))
			}
		}
	}

	// Join with '|' (OR) for discovery search
	flexibleQueryStr := strings.Join(tsQueryParts, " | ")
	if flexibleQueryStr == "" {
		return nil, nil
	}

	searchSQL := `to_tsvector('english', 
		concat_ws(' ', 
			coalesce(disease_name, ''), 
			coalesce(symptoms, ''), 
			coalesce(description, ''), 
			coalesce(advice, '')
		)
	) @@ to_tsquery('english', ?)`

	// Rank by Title matches to ensure relevant clinical guidelines appear first
	err := r.db.Where(searchSQL, flexibleQueryStr).
		Order(gorm.Expr("ts_rank(to_tsvector('english', coalesce(disease_name, '')), to_tsquery('english', ?)) DESC", flexibleQueryStr)).
		Limit(5).Find(&entries).Error
	return entries, err
}

// InitSession creates a blank session row on session start, returning the server-generated ID
func (r *postgresRepository) InitSession(session *models.SymptomSession) error {
	return r.db.Create(session).Error
}

// GetSession fetches a single session and its diagnoses
func (r *postgresRepository) GetSession(sessionID uuid.UUID) (*models.SymptomSession, error) {
	var session models.SymptomSession
	err := r.db.Preload("Diagnoses").First(&session, "id = ?", sessionID).Error
	return &session, err
}

// UpdateSessionChat saves latest chat history JSON and auto-title to Postgres
func (r *postgresRepository) UpdateSessionChat(sessionID uuid.UUID, chatHistoryJSON string, title string) error {
	return r.db.Model(&models.SymptomSession{}).
		Where("id = ?", sessionID).
		Updates(map[string]interface{}{
			"chat_history": chatHistoryJSON,
			"title":        title,
		}).Error
}

// FinalizeSession marks the session as completed and saves diagnosis links
func (r *postgresRepository) FinalizeSession(sessionID uuid.UUID, urgency models.UrgencyLevel, diagnoses []models.Diagnosis) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update the session's urgency level and status
		if err := tx.Model(&models.SymptomSession{}).
			Where("id = ?", sessionID).
			Updates(map[string]interface{}{
				"urgency_level": urgency,
				"status":        "completed",
			}).Error; err != nil {
			return err
		}

		// Create all diagnosis records linked to this session
		if len(diagnoses) > 0 {
			if err := tx.Create(&diagnoses).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ListUserSessions returns all sessions for a specific user, most recent first
func (r *postgresRepository) ListUserSessions(userID uuid.UUID) ([]models.SymptomSession, error) {
	var sessions []models.SymptomSession
	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Select("id, title, status, urgency_level, created_at, updated_at").
		Find(&sessions).Error
	return sessions, err
}
