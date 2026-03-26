package auth

import (
	"context"
	"errors"
	"time"

	"ai-symptom-checker/models"
	"ai-symptom-checker/pkg/email"
	"ai-symptom-checker/pkg/utils"
	"os"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ── DTOs ──────────────────────────────────────────────────────────────────────

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Age      int    `json:"age" binding:"required,min=1,max=120"`
	Gender   string `json:"gender" binding:"required,oneof=male female other"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"` // Optional: If provided, only check that role
}

type DoctorSetupRequest struct {
	InviteToken string `json:"invite_token" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Password    string `json:"password" binding:"required,min=8"`
}

type AdminSetupRequest struct {
	InviteToken string `json:"invite_token" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Username    string `json:"username" binding:"required,min=4"`
	Password    string `json:"password" binding:"required,min=8"`
}

type CompleteWizardRequest struct {
	Specialization string `json:"specialization" binding:"required"`
	Credentials    string `json:"credentials" binding:"required"`
	Bio            string `json:"bio"`
}

type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         interface{} `json:"user"`
	Role         string      `json:"role"`
}

// ── Service Methods ───────────────────────────────────────────────────────────

// RegisterPatient creates a new patient account
func (s *Service) RegisterPatient(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Check if email already exists
	existing, _ := s.repo.FindUserByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("an account with this email already exists")
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	// Generate verification token
	token := utils.GenerateRandomToken()
	expiry := time.Now().Add(24 * time.Hour)

	user := &models.User{
		Name:              req.Name,
		Email:             req.Email,
		Password:          hashed,
		Age:               req.Age,
		Gender:            req.Gender,
		VerificationToken: token,
		VerificationExpiry: &expiry,
		IsEmailVerified:    false,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, errors.New("failed to create account")
	}

	// Send verification email (non-blocking)
	verificationLink := os.Getenv("FRONTEND_URL") + "/verify?token=" + token
	go email.SendVerificationEmail(user.Name, user.Email, verificationLink)

	// We return empty codes to signal that the user is NOT logged in yet
	return &AuthResponse{User: user, Role: "patient"}, nil
}

// Login authenticates a user/doctor/admin
func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	// If a specific role is requested, only try that
	if req.Role != "" {
		switch req.Role {
		case "patient":
			return s.loginPatient(ctx, req)
		case "doctor":
			return s.loginDoctor(ctx, req)
		case "admin":
			return s.loginAdmin(ctx, req)
		default:
			return nil, errors.New("invalid role specified")
		}
	}

	// Automatic detection: Try patient, doctor, and admin in parallel to avoid sequential timeouts
	type result struct {
		res *AuthResponse
		err error
	}
	resChan := make(chan result, 3)

	go func() {
		res, err := s.loginPatient(ctx, req)
		resChan <- result{res, err}
	}()
	go func() {
		res, err := s.loginDoctor(ctx, req)
		resChan <- result{res, err}
	}()
	go func() {
		res, err := s.loginAdmin(ctx, req)
		resChan <- result{res, err}
	}()

	var firstErr error
	for i := 0; i < 3; i++ {
		r := <-resChan
		if r.err == nil {
			return r.res, nil
		}
		if firstErr == nil {
			firstErr = r.err
		}
	}

	return nil, errors.New("invalid email or password")
}

func (s *Service) loginPatient(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}
	if !user.IsActive {
		return nil, errors.New("account is suspended")
	}
	if !user.IsEmailVerified {
		return nil, errors.New("please verify your email before logging in")
	}
	if !utils.ComparePassword(user.Password, req.Password) {
		return nil, errors.New("invalid email or password")
	}
	accessToken, err := utils.GenerateToken(user.ID, user.Email, "patient", false)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}
	refreshToken, err := utils.GenerateToken(user.ID, user.Email, "patient", true)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}
	return &AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: user, Role: "patient"}, nil
}

func (s *Service) loginDoctor(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	doctor, err := s.repo.FindDoctorByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}
	if doctor.Status != models.DoctorStatusActive {
		return nil, errors.New("doctor account is not active")
	}
	if !utils.ComparePassword(doctor.Password, req.Password) {
		return nil, errors.New("invalid email or password")
	}
	accessToken, err := utils.GenerateToken(doctor.ID, doctor.Email, "doctor", false)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}
	refreshToken, err := utils.GenerateToken(doctor.ID, doctor.Email, "doctor", true)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}
	return &AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: doctor, Role: "doctor"}, nil
}

func (s *Service) loginAdmin(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	admin, err := s.repo.FindAdminByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if !admin.IsActive {
		return nil, errors.New("admin account is inactive")
	}
	if !utils.ComparePassword(admin.Password, req.Password) {
		return nil, errors.New("invalid credentials")
	}
	accessToken, err := utils.GenerateToken(admin.ID, admin.Email, "admin", false)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}
	refreshToken, err := utils.GenerateToken(admin.ID, admin.Email, "admin", true)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}
	return &AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: admin, Role: "admin"}, nil
}

// RefreshTokens handles access token renewal via refresh token validation
func (s *Service) RefreshTokens(ctx context.Context, refreshTokenStr string) (*AuthResponse, error) {
	claims, err := utils.ValidateRefreshToken(refreshTokenStr)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// Double check user still exists and is active
	user, err := s.GetMe(ctx, claims.UserID, claims.Role)
	if err != nil {
		return nil, errors.New("user no longer exists")
	}

	accessToken, err := utils.GenerateToken(claims.UserID, claims.Email, claims.Role, false)
	if err != nil {
		return nil, errors.New("failed to generate new access token")
	}

	// Optional: Rotate refresh token too
	newRefreshToken, err := utils.GenerateToken(claims.UserID, claims.Email, claims.Role, true)
	if err != nil {
		return nil, errors.New("failed to generate new refresh token")
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         user,
		Role:         claims.Role,
	}, nil
}

// VerifyEmail validates the token and activates the user account
func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	user, err := s.repo.FindUserByVerificationToken(ctx, token)
	if err != nil {
		return errors.New("invalid or expired verification link")
	}

	if user.IsEmailVerified {
		return nil // already verified
	}

	if user.VerificationExpiry != nil && time.Now().After(*user.VerificationExpiry) {
		return errors.New("verification link has expired")
	}

	user.IsEmailVerified = true
	user.VerificationToken = ""
	user.VerificationExpiry = nil

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return errors.New("failed to verify email")
	}

	// Send welcome email
	go email.SendWelcomeEmail(user.Name, user.Email)

	return nil
}

// Logout currently a placeholder, but could be used to blacklist tokens
func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	return nil
}

// SetupDoctorAccount completes doctor onboarding via invite link
func (s *Service) SetupDoctorAccount(ctx context.Context, req DoctorSetupRequest) (*AuthResponse, error) {
	doctor, err := s.repo.FindDoctorByInviteToken(ctx, req.InviteToken)
	if err != nil {
		return nil, errors.New("invalid or expired invite link")
	}

	// Check token expiry
	if doctor.InviteExpiry != nil && time.Now().After(*doctor.InviteExpiry) {
		return nil, errors.New("invite link has expired")
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	// Complete setup (Activation stage)
	doctor.Name = req.Name
	doctor.Password = hashed
	doctor.Status = models.DoctorStatusActive
	doctor.SetupComplete = false // Must do wizard after login
	doctor.InviteToken = ""      // Invalidate token after use
	doctor.InviteExpiry = nil

	if err := s.repo.UpdateDoctor(ctx, doctor); err != nil {
		return nil, errors.New("failed to complete account setup")
	}

	// Send confirmation email
	go email.SendDoctorAccountSetupConfirmation(doctor.Name, doctor.Email)

	accessToken, err := utils.GenerateToken(doctor.ID, doctor.Email, "doctor", false)
	if err != nil {
		return nil, errors.New("setup complete but failed to generate access token")
	}

	refreshToken, err := utils.GenerateToken(doctor.ID, doctor.Email, "doctor", true)
	if err != nil {
		return nil, errors.New("setup complete but failed to generate refresh token")
	}

	return &AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: doctor, Role: "doctor"}, nil
}

// SetupAdminAccount completes admin onboarding via invite link
func (s *Service) SetupAdminAccount(ctx context.Context, req AdminSetupRequest) (*AuthResponse, error) {
	admin, err := s.repo.FindAdminByInviteToken(ctx, req.InviteToken)
	if err != nil {
		return nil, errors.New("invalid or expired invite link")
	}

	if admin.InviteExpiry != nil && time.Now().After(*admin.InviteExpiry) {
		return nil, errors.New("invite link has expired")
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	admin.Name = req.Name
	admin.Username = req.Username
	admin.Password = hashed
	admin.Status = models.AdminStatusActive
	admin.IsActive = true
	admin.InviteToken = ""
	admin.InviteExpiry = nil

	if err := s.repo.UpdateAdmin(ctx, admin); err != nil {
		return nil, errors.New("failed to complete administrator registration")
	}

	accessToken, err := utils.GenerateToken(admin.ID, admin.Email, "admin", false)
	refreshToken, err := utils.GenerateToken(admin.ID, admin.Email, "admin", true)

	return &AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: admin, Role: "admin"}, nil
}

// CompleteDoctorWizard updates the doctor profile with medical information
func (s *Service) CompleteDoctorWizard(ctx context.Context, doctorID uuid.UUID, req CompleteWizardRequest) error {
	doctor, err := s.repo.FindDoctorByID(ctx, doctorID)
	if err != nil {
		return errors.New("doctor not found")
	}

	doctor.Specialization = req.Specialization
	doctor.Credentials = req.Credentials
	doctor.Bio = req.Bio
	doctor.SetupComplete = true

	if err := s.repo.UpdateDoctor(ctx, doctor); err != nil {
		return errors.New("failed to save profile")
	}

	return nil
}

// GetMe returns the currently authenticated user's profile
func (s *Service) GetMe(ctx context.Context, userID uuid.UUID, role string) (interface{}, error) {
	switch role {
	case "patient":
		return s.repo.FindUserByID(ctx, userID)
	case "doctor":
		return s.repo.FindDoctorByID(ctx, userID)
	case "admin":
		return s.repo.FindAdminByID(ctx, userID)
	default:
		return nil, errors.New("unsupported role")
	}
}

// Needed by auth.repository.go — exported sentinel errors
var _ = utils.ErrNotFound // ensure it's referenced

func init() {
	_ = gorm.ErrRecordNotFound // silence unused import
}
