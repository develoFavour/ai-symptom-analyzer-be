package auth

import (
	"ai-symptom-checker/pkg/middleware"
	"ai-symptom-checker/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body", err)
		return
	}

	res, err := h.service.RegisterPatient(c.Request.Context(), req)
	if err != nil {
		utils.BadRequest(c, err.Error(), err)
		return
	}

	// If we got access tokens, set cookies (shouldn't happen with email verification on)
	if res.AccessToken != "" {
		setAuthCookies(c, res.AccessToken, res.RefreshToken, res.Role)
	}

	utils.Created(c, "Verification email sent. Please check your inbox.", res)
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.BadRequest(c, "Verification token is required", nil)
		return
	}

	err := h.service.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		utils.BadRequest(c, err.Error(), err)
		return
	}

	utils.Ok(c, "Email verified successfully. You can now log in.", nil)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body", err)
		return
	}

	res, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}

	setAuthCookies(c, res.AccessToken, res.RefreshToken, res.Role)
	utils.Ok(c, "Login successful", res)
}

func (h *Handler) SetupDoctorAccount(c *gin.Context) {
	var req DoctorSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body", err)
		return
	}

	res, err := h.service.SetupDoctorAccount(c.Request.Context(), req)
	if err != nil {
		utils.BadRequest(c, err.Error(), err)
		return
	}

	setAuthCookies(c, res.AccessToken, res.RefreshToken, res.Role)
	utils.Ok(c, "Doctor account setup complete", res)
}

func (h *Handler) SetupAdminAccount(c *gin.Context) {
	var req AdminSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body", err)
		return
	}

	res, err := h.service.SetupAdminAccount(c.Request.Context(), req)
	if err != nil {
		utils.BadRequest(c, err.Error(), err)
		return
	}

	setAuthCookies(c, res.AccessToken, res.RefreshToken, res.Role)
	utils.Ok(c, "Admin account setup complete", res)
}

func (h *Handler) CompleteDoctorWizard(c *gin.Context) {
	var req CompleteWizardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body", err)
		return
	}

	userID, _ := c.Get(middleware.ContextKeyUserID)
	err := h.service.CompleteDoctorWizard(c.Request.Context(), userID.(uuid.UUID), req)
	if err != nil {
		utils.BadRequest(c, err.Error(), err)
		return
	}

	utils.Ok(c, "Wizard setup complete", nil)
}

func (h *Handler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		utils.Unauthorized(c, "Refresh token missing")
		return
	}

	res, err := h.service.RefreshTokens(c.Request.Context(), refreshToken)
	if err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}

	setAuthCookies(c, res.AccessToken, res.RefreshToken, res.Role)
	utils.Ok(c, "Tokens refreshed", res)
}

func (h *Handler) Logout(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	if userID != nil {
		h.service.Logout(c.Request.Context(), userID.(uuid.UUID))
	}

	clearAuthCookies(c)
	utils.Ok(c, "Logged out successfully", nil)
}

func (h *Handler) GetMe(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	role, _ := c.Get(middleware.ContextKeyRole)

	user, err := h.service.GetMe(c.Request.Context(), userID.(uuid.UUID), role.(string))
	if err != nil {
		utils.NotFound(c, "User not found")
		return
	}

	utils.Ok(c, "Profile retrieved", user)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func setAuthCookies(c *gin.Context, access, refresh, role string) {
	// Access token cookie (shorter life)
	c.SetCookie("access_token", access, int(15*60), "/", "", false, true) // 15 mins
	// Refresh token cookie (longer life)
	c.SetCookie("refresh_token", refresh, int(7*24*60*60), "/", "", false, true) // 7 days
	// Role cookie for middleware (shorter life matching access)
	c.SetCookie("user_role", role, int(15*60), "/", "", false, true)
}

func clearAuthCookies(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.SetCookie("user_role", "", -1, "/", "", false, true)
}
