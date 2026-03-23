package admin

import (
    "context"

    "ai-symptom-checker/models"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Repository interface {
	CreateDoctor(ctx context.Context, doctor *models.Doctor) error
	FindDoctorByEmail(ctx context.Context, email string) (*models.Doctor, error)
	GetPatientCount(ctx context.Context) (int64, error)
	GetActiveDoctorCount(ctx context.Context) (int64, error)
	GetPendingInviteCount(ctx context.Context) (int64, error)
	GetTotalSymptomSessions(ctx context.Context) (int64, error)
	ListDoctors(ctx context.Context) ([]models.Doctor, error)
	UpdateDoctorStatus(ctx context.Context, id uuid.UUID, status models.DoctorStatus) error
	ListUsers(ctx context.Context) ([]models.User, error)
	UpdateUserStatus(ctx context.Context, id uuid.UUID, isActive bool) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	
	// Admin Management
	CreateAdmin(ctx context.Context, admin *models.Admin) error
	ListAdmins(ctx context.Context) ([]models.Admin, error)
	FindAdminByEmail(ctx context.Context, email string) (*models.Admin, error)
	FindAdminByInviteToken(ctx context.Context, token string) (*models.Admin, error)
	UpdateAdmin(ctx context.Context, admin *models.Admin) error
	DeleteAdmin(ctx context.Context, id uuid.UUID) error
	UnscopedDeleteAdmin(ctx context.Context, id uuid.UUID) error

	LogAction(ctx context.Context, log *models.AdminLog) error
	GetRecentLogs(ctx context.Context, limit int) ([]models.AdminLog, error)

	// Analytics
	GetDiagnosticActivity(ctx context.Context, days int) ([]map[string]interface{}, error)
	GetEpidemicDistribution(ctx context.Context) ([]map[string]interface{}, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) LogAction(ctx context.Context, log *models.AdminLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *repository) GetRecentLogs(ctx context.Context, limit int) ([]models.AdminLog, error) {
	var logs []models.AdminLog
	err := r.db.WithContext(ctx).
		Preload("Admin").
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *repository) CreateDoctor(ctx context.Context, doctor *models.Doctor) error {
	return r.db.WithContext(ctx).Create(doctor).Error
}

func (r *repository) FindDoctorByEmail(ctx context.Context, email string) (*models.Doctor, error) {
	var doctor models.Doctor
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&doctor).Error
	if err != nil {
		return nil, err
	}
	return &doctor, nil
}

func (r *repository) GetPatientCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error
	return count, err
}

func (r *repository) GetActiveDoctorCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Doctor{}).Where("status = ?", models.DoctorStatusActive).Count(&count).Error
	return count, err
}

func (r *repository) GetPendingInviteCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Doctor{}).Where("status = ?", models.DoctorStatusPending).Count(&count).Error
	return count, err
}

func (r *repository) GetTotalSymptomSessions(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.SymptomSession{}).Count(&count).Error
	return count, err
}

func (r *repository) ListDoctors(ctx context.Context) ([]models.Doctor, error) {
	var doctors []models.Doctor
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&doctors).Error
	return doctors, err
}

func (r *repository) UpdateDoctorStatus(ctx context.Context, id uuid.UUID, status models.DoctorStatus) error {
	return r.db.WithContext(ctx).Model(&models.Doctor{}).Where("id = ?", id).Update("status", status).Error
}

func (r *repository) ListUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&users).Error
	return users, err
}

func (r *repository) UpdateUserStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("is_active", isActive).Error
}

func (r *repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

// ── Admin Management ──────────────────────────────────────────────────────────

func (r *repository) CreateAdmin(ctx context.Context, admin *models.Admin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *repository) ListAdmins(ctx context.Context) ([]models.Admin, error) {
	var admins []models.Admin
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&admins).Error
	return admins, err
}

func (r *repository) FindAdminByEmail(ctx context.Context, email string) (*models.Admin, error) {
	var admin models.Admin
	// Use Unscoped to find even soft-deleted admins to prevent unique constraint violations
	err := r.db.WithContext(ctx).Unscoped().Where("email = ?", email).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *repository) FindAdminByInviteToken(ctx context.Context, token string) (*models.Admin, error) {
	var admin models.Admin
	err := r.db.WithContext(ctx).Where("invite_token = ?", token).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *repository) UpdateAdmin(ctx context.Context, admin *models.Admin) error {
	return r.db.WithContext(ctx).Save(admin).Error
}

func (r *repository) DeleteAdmin(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Admin{}, id).Error
}

func (r *repository) UnscopedDeleteAdmin(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&models.Admin{}, id).Error
}

func (r *repository) GetDiagnosticActivity(ctx context.Context, days int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	// Group by day truncated to date
	query := `SELECT 
				TO_CHAR(created_at, 'Dy') as name, 
				COUNT(*) as assessments,
				AVG(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) * 100 as accuracy
			  FROM symptom_sessions 
			  WHERE created_at > NOW() - INTERVAL '7 days'
			  GROUP BY TO_CHAR(created_at, 'Dy'), date_trunc('day', created_at)
			  ORDER BY date_trunc('day', created_at) ASC`
	
	err := r.db.WithContext(ctx).Raw(query).Scan(&results).Error
	return results, err
}

func (r *repository) GetEpidemicDistribution(ctx context.Context) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	// Pull top 5 diagnosed conditions from the last 30 days
	query := `SELECT 
				condition_name as name, 
				COUNT(*) as count
			  FROM diagnoses 
			  WHERE created_at > NOW() - INTERVAL '30 days'
			  GROUP BY condition_name
			  ORDER BY count DESC
			  LIMIT 5`
	
	err := r.db.WithContext(ctx).Raw(query).Scan(&results).Error
	return results, err
}
