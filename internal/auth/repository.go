package auth

import (
	"context"
	"errors"

	"ai-symptom-checker/models"
	"ai-symptom-checker/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines the DB contract for the auth module
type Repository interface {
	CreateUser(ctx context.Context, user *models.User) error
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindAdminByEmail(ctx context.Context, email string) (*models.Admin, error)
	FindAdminByID(ctx context.Context, id uuid.UUID) (*models.Admin, error)
	FindDoctorByEmail(ctx context.Context, email string) (*models.Doctor, error)
	FindDoctorByID(ctx context.Context, id uuid.UUID) (*models.Doctor, error)
	FindDoctorByInviteToken(ctx context.Context, token string) (*models.Doctor, error)
	FindUserByVerificationToken(ctx context.Context, token string) (*models.User, error)
	FindAdminByInviteToken(ctx context.Context, token string) (*models.Admin, error)
	UpdateUser(ctx context.Context, user *models.User) error
	UpdateDoctor(ctx context.Context, doctor *models.Doctor) error
	UpdateAdmin(ctx context.Context, admin *models.Admin) error
}

// postgresRepository is the concrete GORM implementation
type postgresRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *postgresRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrNotFound
	}
	return &user, err
}

func (r *postgresRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrNotFound
	}
	return &user, err
}

func (r *postgresRepository) FindAdminByID(ctx context.Context, id uuid.UUID) (*models.Admin, error) {
	var admin models.Admin
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrNotFound
	}
	return &admin, err
}

func (r *postgresRepository) FindAdminByEmail(ctx context.Context, email string) (*models.Admin, error) {
	var admin models.Admin
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrNotFound
	}
	return &admin, err
}

func (r *postgresRepository) FindDoctorByEmail(ctx context.Context, email string) (*models.Doctor, error) {
	var doctor models.Doctor
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&doctor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrNotFound
	}
	return &doctor, err
}

func (r *postgresRepository) FindDoctorByID(ctx context.Context, id uuid.UUID) (*models.Doctor, error) {
	var doctor models.Doctor
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&doctor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrNotFound
	}
	return &doctor, err
}

func (r *postgresRepository) FindDoctorByInviteToken(ctx context.Context, token string) (*models.Doctor, error) {
	var doctor models.Doctor
	err := r.db.WithContext(ctx).Where("invite_token = ?", token).First(&doctor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrNotFound
	}
	return &doctor, err
}

func (r *postgresRepository) FindUserByVerificationToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("verification_token = ?", token).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrNotFound
	}
	return &user, err
}

func (r *postgresRepository) FindAdminByInviteToken(ctx context.Context, token string) (*models.Admin, error) {
	var admin models.Admin
	err := r.db.WithContext(ctx).Where("invite_token = ?", token).First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrNotFound
	}
	return &admin, err
}

func (r *postgresRepository) UpdateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *postgresRepository) UpdateDoctor(ctx context.Context, doctor *models.Doctor) error {
	return r.db.WithContext(ctx).Save(doctor).Error
}

func (r *postgresRepository) UpdateAdmin(ctx context.Context, admin *models.Admin) error {
	return r.db.WithContext(ctx).Save(admin).Error
}
