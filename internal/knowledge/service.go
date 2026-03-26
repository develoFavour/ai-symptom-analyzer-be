package knowledge

import (
	"ai-symptom-checker/models"
	"ai-symptom-checker/pkg/ai"
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
)

type CreateEntryRequest struct {
	Title           string                   `json:"title" binding:"required"`
	Source          string                   `json:"source" binding:"required"`
	Category        models.KnowledgeCategory `json:"category" binding:"required"`
	Content         string                   `json:"content" binding:"required"`
	Tags            string                   `json:"tags"`
	IsEpidemicAlert bool                     `json:"is_epidemic_alert"`
	Region          string                   `json:"region"`
	UrgencyScore    int                      `json:"urgency_score"`
}

const (
	placeholderDerived        = "Derived from symptom knowledge graph associations."
	placeholderHuggingFace    = "Information loaded from ML disease dataset."
	placeholderPatientDerived = "Symptoms descriptions manually compiled from patient data."
)

type Service struct {
	repo     Repository
	aiClient ai.Client
}

func NewService(repo Repository, aiClient ai.Client) *Service {
	return &Service{repo: repo, aiClient: aiClient}
}

func (s *Service) CreateEntry(ctx context.Context, adminID uuid.UUID, req CreateEntryRequest) (*models.KnowledgeEntry, error) {
	entry := &models.KnowledgeEntry{
		Title:           req.Title,
		Source:          req.Source,
		Category:        req.Category,
		Description:     req.Content, // Map Content from req to Description in DB
		Symptoms:        "",          // Default empty for now, or could map from req
		Tags:            req.Tags,
		IsEpidemicAlert: req.IsEpidemicAlert,
		Region:          req.Region,
		UrgencyScore:    req.UrgencyScore,
		AuthorID:        &adminID,
		Status:          models.KnowledgeStatusActive,
	}

	if err := s.repo.CreateEntry(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *Service) ListEntries(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]models.KnowledgeEntry, int64, error) {
	return s.repo.ListEntries(ctx, filter, limit, offset)
}

func (s *Service) InquireEntries(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]models.KnowledgeEntry, int64, error) {
	// For doctors, we only show active records
	filter["status"] = models.KnowledgeStatusActive
	return s.repo.ListEntries(ctx, filter, limit, offset)
}

func (s *Service) enrichDescriptions(ctx context.Context, entries []models.KnowledgeEntry) {
	for i := range entries {
		desc := entries[i].Description
		if desc == placeholderDerived || desc == placeholderHuggingFace || desc == placeholderPatientDerived {
			// Generate and persist
			newDesc, err := s.aiClient.DefineMedicalTerm(ctx, entries[i].Title)
			if err != nil {
				log.Printf("[KnowledgeService] AI enrichment failed for %s: %v", entries[i].Title, err)
				continue
			}
			entries[i].Description = newDesc
			// Update in background to avoid blocking user too long (though we already generated it)
			go func(entry models.KnowledgeEntry) {
				if err := s.repo.UpdateEntry(context.Background(), &entry); err != nil {
					log.Printf("[KnowledgeService] Failed to persist AI description for %s: %v", entry.Title, err)
				}
			}(entries[i])
		}
	}
}

func (s *Service) UpdateEntry(ctx context.Context, id uuid.UUID, req CreateEntryRequest) error {
	entry, err := s.repo.GetEntryByID(ctx, id)
	if err != nil {
		return errors.New("entry not found")
	}

	entry.Title = req.Title
	entry.Source = req.Source
	entry.Category = req.Category
	entry.Description = req.Content
	entry.Tags = req.Tags
	entry.IsEpidemicAlert = req.IsEpidemicAlert
	entry.Region = req.Region
	entry.UrgencyScore = req.UrgencyScore

	return s.repo.UpdateEntry(ctx, entry)
}

func (s *Service) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteEntry(ctx, id)
}

func (s *Service) SearchVerifiedEvidence(ctx context.Context, query string) ([]models.KnowledgeEntry, error) {
	return s.repo.SearchSimilar(ctx, query, 3)
}

func (s *Service) GetGlobalThreats(ctx context.Context) ([]models.KnowledgeEntry, error) {
	return s.repo.GetActiveEpidemicAlerts(ctx)
}
