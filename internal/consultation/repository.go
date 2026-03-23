package consultation

import (
	"ai-symptom-checker/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines the DB contract for the consultation module
type Repository interface {
	// Doctor discovery
	ListActiveDoctors() ([]models.Doctor, error)

	// Consultation CRUD
	Create(c *models.Consultation) error
	GetByID(id uuid.UUID) (*models.Consultation, error)
	ListByPatient(userID uuid.UUID) ([]models.Consultation, error)
	ListByDoctor(doctorID uuid.UUID) ([]models.Consultation, error)

	// Session linkage (for fetching chat history for summary)
	GetSessionChatHistory(sessionID uuid.UUID) (string, error)

	// Actions
	CreateReply(reply *models.ConsultationReply) error
	CreateMessage(m *models.ConsultationMessage) error
	UpdateStatus(consultationID uuid.UUID, status models.ConsultationStatus) error
}

type postgresRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) ListActiveDoctors() ([]models.Doctor, error) {
	var doctors []models.Doctor
	err := r.db.
		Where("status = ? AND is_active = ?", models.DoctorStatusActive, true).
		Select("id, name, email, specialization, credentials, created_at").
		Find(&doctors).Error
	return doctors, err
}

func (r *postgresRepository) Create(c *models.Consultation) error {
	return r.db.Create(c).Error
}

func (r *postgresRepository) GetByID(id uuid.UUID) (*models.Consultation, error) {
	var c models.Consultation
	err := r.db.Preload("User").Preload("Doctor").Preload("Replies.Doctor").Preload("Messages").First(&c, "id = ?", id).Error
	return &c, err
}

func (r *postgresRepository) ListByPatient(userID uuid.UUID) ([]models.Consultation, error) {
	var list []models.Consultation
	err := r.db.
		Where("user_id = ?", userID).
		Preload("Doctor").
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *postgresRepository) ListByDoctor(doctorID uuid.UUID) ([]models.Consultation, error) {
	var list []models.Consultation
	err := r.db.
		Where("doctor_id = ?", doctorID).
		Preload("User").
		Preload("Replies").
		Preload("Session.Diagnoses").
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *postgresRepository) GetSessionChatHistory(sessionID uuid.UUID) (string, error) {
	var session models.SymptomSession
	err := r.db.Select("chat_history").First(&session, "id = ?", sessionID).Error
	return session.ChatHistory, err
}

func (r *postgresRepository) CreateReply(reply *models.ConsultationReply) error {
	return r.db.Create(reply).Error
}

func (r *postgresRepository) CreateMessage(m *models.ConsultationMessage) error {
	return r.db.Create(m).Error
}

func (r *postgresRepository) UpdateStatus(id uuid.UUID, status models.ConsultationStatus) error {
	return r.db.Model(&models.Consultation{}).Where("id = ?", id).Update("status", status).Error
}
