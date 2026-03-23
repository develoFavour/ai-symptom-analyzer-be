package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

const brevoAPIURL = "https://api.brevo.com/v3/smtp/email"

type Contact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type brevoPayload struct {
	Sender      Contact   `json:"sender"`
	To          []Contact `json:"to"`
	Subject     string    `json:"subject"`
	HTMLContent string    `json:"htmlContent"`
}

// send is the core Brevo dispatch function
func send(to Contact, subject, htmlContent string) error {
	payload := brevoPayload{
		Sender: Contact{
			Name:  os.Getenv("BREVO_SENDER_NAME"),
			Email: os.Getenv("BREVO_SENDER_EMAIL"),
		},
		To:          []Contact{to},
		Subject:     subject,
		HTMLContent: htmlContent,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[Email] Marshal error: %v", err)
		return fmt.Errorf("brevo: marshal error: %w", err)
	}

	req, err := http.NewRequest("POST", brevoAPIURL, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[Email] Request creation error: %v", err)
		return fmt.Errorf("brevo: request creation error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", os.Getenv("BREVO_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Email] Dispatch error to %s: %v", to.Email, err)
		return fmt.Errorf("brevo: send error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		log.Printf("[Email] Brevo API error (%d) for %s: %v", resp.StatusCode, to.Email, errResp)
		return fmt.Errorf("brevo: API error status %d", resp.StatusCode)
	}

	log.Printf("[Email] Successfully sent '%s' to %s", subject, to.Email)
	return nil
}

// ── Email Templates ───────────────────────────────────────────────────────────

// SendVerificationEmail — sent when a new patient registers
func SendVerificationEmail(name, email, verificationLink string) error {
	html := fmt.Sprintf(`
	<div style="font-family:sans-serif;max-width:600px;margin:auto;padding:20px;border:1px solid #e5e7eb;border-radius:12px">
		<h2 style="color:#0a2a2a">Verify Your Email Address 📧</h2>
		<p>Hello %s,</p>
		<p>Thank you for signing up for AI Symptom Checker. To complete your registration and start using our platform, please verify your email address by clicking the button below:</p>
		<div style="text-align:center;margin:30px 0">
			<a href="%s" style="background:#0a2a2a;color:#fff;padding:14px 28px;border-radius:50px;text-decoration:none;display:inline-block;font-weight:bold shadow:0 10px 15px -3px rgba(0,0,0,0.1)">
				Verify Email Address
			</a>
		</div>
		<p style="color:#6b7280;font-size:14px">This link will expire in 24 hours. If you didn't create an account, you can safely ignore this email.</p>
		<hr style="border:0;border-top:1px solid #e5e7eb;margin:20px 0" />
		<p style="color:#9ca3af;font-size:12px">
			<strong>Disclaimer:</strong> This tool provides preliminary health assessments. Always consult a medical professional for actual advice.
		</p>
	</div>`, name, verificationLink)
	return send(Contact{Name: name, Email: email}, "Verify Your Email - AI Symptom Checker", html)
}

// SendWelcomeEmail — sent after successful verification
func SendWelcomeEmail(name, email string) error {
	html := fmt.Sprintf(`
	<div style="font-family:sans-serif;max-width:600px;margin:auto;padding:20px;border:1px solid #e5e7eb;border-radius:12px">
		<h2 style="color:#0a2a2a">Welcome aboard, %s! 🎊</h2>
		<p>Your email has been successfully verified. Your account is now fully active.</p>
		<p>You can now log in to use our AI-powered symptom analysis and consult with specialists.</p>
		<div style="text-align:center;margin:30px 0">
			<a href="%s/login" style="background:#0a2a2a;color:#fff;padding:14px 28px;border-radius:50px;text-decoration:none;display:inline-block;font-weight:bold">
				Login to Dashboard
			</a>
		</div>
	</div>`, name, os.Getenv("FRONTEND_URL"))
	return send(Contact{Name: name, Email: email}, "Welcome to AI Symptom Checker", html)
}

// SendDoctorInviteEmail — admin invites a doctor (ONLY WAY doctors enter the system)
func SendDoctorInviteEmail(name, email, inviteLink string) error {
	html := fmt.Sprintf(`
	<div style="font-family:sans-serif;max-width:600px;margin:auto">
		<h2 style="color:#2563eb">You've Been Invited to AI Symptom Checker 🩺</h2>
		<p>Hello Dr. %s,</p>
		<p>You have been invited by our admin team to join the platform as a healthcare expert.</p>
		<p>
			<a href="%s" style="background:#2563eb;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;display:inline-block">
				Set Up Your Account
			</a>
		</p>
		<p style="color:#6b7280;font-size:13px">This invitation link expires in <strong>48 hours</strong>.</p>
	</div>`, name, inviteLink)
	return send(Contact{Name: name, Email: email}, "Invitation to Join AI Symptom Checker", html)
}

// SendDoctorAccountSetupConfirmation — sent when doctor completes account setup via invite link
func SendDoctorAccountSetupConfirmation(name, email string) error {
	html := fmt.Sprintf(`
	<div style="font-family:sans-serif;max-width:600px;margin:auto">
		<h2 style="color:#16a34a">Account Setup Complete ✅</h2>
		<p>Hello Dr. %s,</p>
		<p>Your doctor account has been set up successfully. You can now log in to access your dashboard,
		   review patient consultation requests, and contribute to the medical knowledge base.</p>
	</div>`, name)
	return send(Contact{Name: name, Email: email}, "Your Doctor Account Is Ready", html)
}

// SendDoctorAccessRevokedEmail — sent when admin suspends/revokes a doctor
func SendDoctorAccessRevokedEmail(name, email string) error {
	html := fmt.Sprintf(`
	<div style="font-family:sans-serif;max-width:600px;margin:auto">
		<h2 style="color:#dc2626">Account Access Revoked</h2>
		<p>Hello Dr. %s,</p>
		<p>Your access to the AI Symptom Checker platform has been revoked by the administrator.</p>
		<p>If you believe this is an error, please contact our support team.</p>
	</div>`, name)
	return send(Contact{Name: name, Email: email}, "Your Access Has Been Revoked", html)
}

// SendAdminInviteEmail — root/existing admin invites a new admin
func SendAdminInviteEmail(email, inviteLink string) error {
	html := fmt.Sprintf(`
	<div style="font-family:sans-serif;max-width:600px;margin:auto;padding:24px;border:1px solid #1e293b;border-radius:16px">
		<h2 style="color:#1e293b">Administrative Invitation 🛡️</h2>
		<p>Hello,</p>
		<p>You have been invited to join the <strong>Vitalis AI</strong> administrative governance team.</p>
		<p>As an administrator, you will have complete oversight of system telemetry, medical staff moderation, and knowledge base integrity.</p>
		<div style="margin:32px 0;text-align:center">
			<a href="%s" style="background:#1e293b;color:#ffffff;padding:16px 32px;border-radius:8px;text-decoration:none;display:inline-block;font-weight:bold">
				Complete Admin Setup
			</a>
		</div>
		<p style="color:#64748b;font-size:13px">This secure invitation link expires in <strong>24 hours</strong>. If you did not expect this invitation, please notify the security lead immediately.</p>
	</div>`, inviteLink)
	return send(Contact{Name: "Administrator", Email: email}, "Administrative Invitation - Vitalis AI", html)
}

// SendAdminAccountSetupConfirmation — sent when admin completes setup
func SendAdminAccountSetupConfirmation(name, email string) error {
	html := fmt.Sprintf(`
	<div style="font-family:sans-serif;max-width:600px;margin:auto;padding:24px;border:1px solid #10b981;border-radius:16px">
		<h2 style="color:#10b981">Admin Account Active ✅</h2>
		<p>Hello %s,</p>
		<p>Your administrative account setup is complete. You now have full executive access to the Vitalis AI platform.</p>
		<p>Please ensure Multi-Factor Authentication is enabled on your first login for HIPAA compliance.</p>
	</div>`, name)
	return send(Contact{Name: name, Email: email}, "Admin Executive Access Confirmed", html)
}

// SendConsultationResponseEmail — sent to patient when doctor replies
func SendConsultationResponseEmail(patientName, email string) error {
	html := fmt.Sprintf(`
	<div style="font-family:sans-serif;max-width:600px;margin:auto">
		<h2 style="color:#2563eb">A Doctor Has Responded to Your Consultation 💬</h2>
		<p>Hello %s,</p>
		<p>A doctor has reviewed your consultation request and provided a response.</p>
		<p>Log in to your dashboard to view their advice and recommendations.</p>
	</div>`, patientName)
	return send(Contact{Name: patientName, Email: email}, "Doctor Responded to Your Consultation", html)
}

// SendPasswordResetEmail — password reset link
func SendPasswordResetEmail(name, email, resetLink string) error {
	html := fmt.Sprintf(`
	<div style="font-family:sans-serif;max-width:600px;margin:auto">
		<h2 style="color:#2563eb">Password Reset Request 🔐</h2>
		<p>Hello %s,</p>
		<p>We received a request to reset your password.</p>
		<p>
			<a href="%s" style="background:#2563eb;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;display:inline-block">
				Reset My Password
			</a>
		</p>
		<p style="color:#6b7280;font-size:13px">
			This link expires in <strong>1 hour</strong>.
			If you didn't request this, please ignore this email.
		</p>
	</div>`, name, resetLink)
	return send(Contact{Name: name, Email: email}, "Reset Your Password", html)
}
