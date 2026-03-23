package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"ai-symptom-checker/models"
	"ai-symptom-checker/pkg/email"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ── DTOs ──────────────────────────────────────────────────────────────────────

type InviteDoctorRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type InviteAdminRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ── Service Methods ───────────────────────────────────────────────────────────

// InviteDoctor starts the onboarding flow for a new doctor
func (s *Service) InviteDoctor(ctx context.Context, adminID uuid.UUID, req InviteDoctorRequest) error {
	existing, _ := s.repo.FindDoctorByEmail(ctx, req.Email)
	if existing != nil {
		return errors.New("a doctor with this email is already in the system")
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return errors.New("failed to generate secure invitation")
	}
	token := hex.EncodeToString(tokenBytes)

	expiry := time.Now().Add(48 * time.Hour)

	doctor := &models.Doctor{
		Name:             req.Name,
		Email:            req.Email,
		Password:         "PENDING_INVITE",
		InviteToken:      token,
		InviteExpiry:     &expiry,
		Status:           models.DoctorStatusPending,
		InvitedByAdminID: adminID,
	}

	if err := s.repo.CreateDoctor(ctx, doctor); err != nil {
		return errors.New("failed to create doctor invitation")
	}

	_ = s.repo.LogAction(ctx, &models.AdminLog{
		AdminID:  adminID,
		Action:   "invited_doctor",
		TargetID: doctor.ID.String(),
		Details:  "Invited Dr. " + req.Name + " (" + req.Email + ")",
	})

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	inviteLink := frontendURL + "/verify/doctor?token=" + token

	log.Printf("[AdminService] GENREATED invite link for %s: %s", req.Email, inviteLink)

	go func() {
		if err := email.SendDoctorInviteEmail(req.Name, req.Email, inviteLink); err != nil {
			log.Printf("[AdminService] FAILED to send invitation to %s: %v", req.Email, err)
		} else {
			log.Printf("[AdminService] Invitation dispatched successfully to %s", req.Email)
		}
	}()

	return nil
}

// GetStats returns high-level system stats for the admin dashboard
func (s *Service) GetStats(ctx context.Context) (map[string]interface{}, error) {
	patientCount, _ := s.repo.GetPatientCount(ctx)
	doctorCount, _ := s.repo.GetActiveDoctorCount(ctx)
	pendingInvites, _ := s.repo.GetPendingInviteCount(ctx)
	totalSessions, _ := s.repo.GetTotalSymptomSessions(ctx)
	recentLogs, _ := s.repo.GetRecentLogs(ctx, 5)
	activity, _ := s.repo.GetDiagnosticActivity(ctx, 7)
	distribution, _ := s.repo.GetEpidemicDistribution(ctx)

	formattedLogs := []map[string]interface{}{}
	for _, log := range recentLogs {
		userName := "System"
		if log.Admin.Name != "" {
			userName = log.Admin.Name
		}
		formattedLogs = append(formattedLogs, map[string]interface{}{
			"user":   userName,
			"action": log.Details,
			"time":   log.CreatedAt,
		})
	}

	return map[string]interface{}{
		"total_patients":    patientCount,
		"active_doctors":    doctorCount,
		"pending_invites":   pendingInvites,
		"total_assessments": totalSessions,
		"recent_actions":    formattedLogs,
		"activity":          activity,
		"distribution":      distribution,
	}, nil
}

// ── Doctor Management ────────────────────────────────────────────────────────

func (s *Service) ListDoctors(ctx context.Context) ([]models.Doctor, error) {
	return s.repo.ListDoctors(ctx)
}

func (s *Service) UpdateDoctorStatus(ctx context.Context, adminID uuid.UUID, id uuid.UUID, status string) error {
	newStatus := models.DoctorStatus(status)
	err := s.repo.UpdateDoctorStatus(ctx, id, newStatus)
	if err == nil {
		_ = s.repo.LogAction(ctx, &models.AdminLog{
			AdminID:  adminID,
			Action:   "updated_doctor_status",
			TargetID: id.String(),
			Details:  fmt.Sprintf("Changed doctor status to %s", status),
		})
	}
	return err
}

// ── User Management ──────────────────────────────────────────────────────────

func (s *Service) ListUsers(ctx context.Context) ([]models.User, error) {
	return s.repo.ListUsers(ctx)
}

func (s *Service) UpdateUserStatus(ctx context.Context, adminID uuid.UUID, id uuid.UUID, isActive bool) error {
	err := s.repo.UpdateUserStatus(ctx, id, isActive)
	if err == nil {
		statusStr := "suspended"
		if isActive {
			statusStr = "activated"
		}
		_ = s.repo.LogAction(ctx, &models.AdminLog{
			AdminID:  adminID,
			Action:   "updated_user_status",
			TargetID: id.String(),
			Details:  fmt.Sprintf("Changed user account status to %s", statusStr),
		})
	}
	return err
}

func (s *Service) DeleteUser(ctx context.Context, adminID uuid.UUID, id uuid.UUID) error {
	err := s.repo.DeleteUser(ctx, id)
	if err == nil {
		_ = s.repo.LogAction(ctx, &models.AdminLog{
			AdminID:  adminID,
			Action:   "deleted_user",
			TargetID: id.String(),
			Details:  "Permanently removed user account",
		})
	}
	return err
}

// ── Admin Onboarding ─────────────────────────────────────────────────────────

func (s *Service) InviteAdmin(ctx context.Context, inviterID uuid.UUID, req InviteAdminRequest) error {
	// 1. Check if email already exists
	existing, _ := s.repo.FindAdminByEmail(ctx, req.Email)
	if existing != nil {
		if existing.DeletedAt.Valid {
			// Found a soft-deleted record; purged permanently before re-inviting
			_ = s.repo.UnscopedDeleteAdmin(ctx, existing.ID)
		} else {
			return errors.New("an administrator with this email already exists")
		}
	}

	// 2. Generate invite token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)
	expiry := time.Now().Add(24 * time.Hour)

	// 3. Create pending admin
	admin := &models.Admin{
		Email:        req.Email,
		Name:         "Pending Admin", // Placeholder until setup
		Username:     "pending_" + token[:12],
		Password:     "INVITED_PLACEHOLDER",
		Status:       models.AdminStatusPending,
		IsActive:     true,
		InviteToken:  token,
		InviteExpiry: &expiry,
	}

	if err := s.repo.CreateAdmin(ctx, admin); err != nil {
		return errors.New("failed to create administrative invitation")
	}

	// 4. Log the action
	_ = s.repo.LogAction(ctx, &models.AdminLog{
		AdminID: inviterID,
		Action:  "invited_admin",
		Details: fmt.Sprintf("Invited new administrator: %s", req.Email),
	})

	// 5. Send Email
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	inviteLink := fmt.Sprintf("%s/verify/admin?token=%s", frontendURL, token)

	log.Printf("[AdminService] GENERATED administrative enrollment handle for %s: %s", req.Email, inviteLink)

	go func() {
		if err := email.SendAdminInviteEmail(req.Email, inviteLink); err != nil {
			log.Printf("[AdminService] FAILED to dispatch administrative invitation to %s: %v", req.Email, err)
		} else {
			log.Printf("[AdminService] Administrative invitation dispatched successfully to %s", req.Email)
		}
	}()

	return nil
}

func (s *Service) ListAdmins(ctx context.Context) ([]models.Admin, error) {
	return s.repo.ListAdmins(ctx)
}

func (s *Service) DeleteAdmin(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	// 1. Prevent self-deletion if needed, or allow it
	if actorID == id {
		return errors.New("you cannot remove your own administrative access")
	}

	err := s.repo.DeleteAdmin(ctx, id)
	if err == nil {
		_ = s.repo.LogAction(ctx, &models.AdminLog{
			AdminID:  actorID,
			Action:   "deleted_admin",
			TargetID: id.String(),
			Details:  "Revoked administrative access and removed account",
		})
	}
	return err
}
