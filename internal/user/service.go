package user

import (
	"context"

	"time"

	"ai-symptom-checker/models"

	"github.com/google/uuid"
)

type DashboardResponse struct {
	Stats              map[string]interface{} `json:"stats"`
	RecentHistory      []interface{}          `json:"recent_history"`
	ActiveConsultation *models.Consultation   `json:"active_consultation"`
	HealthTip          string                 `json:"health_tip"`
}

var healthTips = []string{
	"Stay hydrated! Drinking at least 8 glasses of water a day helps your body recover faster.",
	"Maintain good posture while working or using digital devices to prevent back and neck pain.",
	"Incorporate at least 30 minutes of moderate exercise into your daily routine for heart health.",
	"Ensure you get 7-9 hours of quality sleep each night for optimal physical and mental recovery.",
	"Regular hand washing with soap and water is one of the best ways to protect yourself from illness.",
	"Fuel your body with a balanced diet rich in fruits, vegetables, lean proteins, and whole grains.",
	"Practice mindfulness or deep breathing exercises for a few minutes daily to manage stress levels.",
	"Limit your intake of processed sugars and salty snacks to maintain healthy blood pressure.",
	"Schedule regular health checkups with your doctor for early detection of potential health issues.",
	"Protect your skin from sun damage by wearing sunscreen and protective clothing outdoors.",
}

// Service holds repository dependency for the user module
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetDashboardData(ctx context.Context, userID uuid.UUID) (*DashboardResponse, error) {
	stats, err := s.repo.GetDashboardStats(userID)
	if err != nil {
		return nil, err
	}

	history, err := s.repo.GetRecentHistory(userID, 5)
	if err != nil {
		return nil, err
	}

	active, err := s.repo.GetActiveConsultation(userID)
	if err != nil {
		return nil, err
	}

	// Choose tip based on day of the year
	tipIndex := time.Now().YearDay() % len(healthTips)
	tip := healthTips[tipIndex]

	return &DashboardResponse{
		Stats:              stats,
		RecentHistory:      history,
		HealthTip:          tip,
		ActiveConsultation: active,
	}, nil
}
