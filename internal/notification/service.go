package notification

import (
	"context"

	"ai-symptom-checker/models"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUserNotifications(ctx context.Context, userID uuid.UUID) ([]models.Notification, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) MarkRead(ctx context.Context, userID uuid.UUID, notifID uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, userID, notifID)
}

func (s *Service) CreateNotification(ctx context.Context, notif *models.Notification) error {
	return s.repo.Create(ctx, notif)
}
