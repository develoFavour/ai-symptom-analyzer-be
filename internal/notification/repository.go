package notification

import (
	"context"

	"ai-symptom-checker/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines the DB contract for the notification module
type Repository interface {
	Create(ctx context.Context, notif *models.Notification) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Notification, error)
	MarkAsRead(ctx context.Context, userID uuid.UUID, notifID uuid.UUID) error
}

type postgresRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, notif *models.Notification) error {
	return r.db.WithContext(ctx).Create(notif).Error
}

func (r *postgresRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ? OR doctor_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

func (r *postgresRepository) MarkAsRead(ctx context.Context, userID uuid.UUID, notifID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ? AND (user_id = ? OR doctor_id = ?)", notifID, userID, userID).
		Update("is_read", true).Error
}
