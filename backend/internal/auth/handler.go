package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp/totp"
	"github.com/username/bulletin/backend/internal/mailer"
	"github.com/username/bulletin/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	DB     *pgxpool.Pool
	Mailer *mailer.Mailer
}

func NewHandler(db *pgxpool.Pool, m *mailer.Mailer) *Handler {
	return &Handler{DB: db, Mailer: m}
}

type registerRequest struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	InviteCode string `json:"invite_code"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Verify invite code
	var inviteID uuid.UUID
	var circleID uuid.UUID
	var invitedByID *uuid.UUID // Use pointer for NULLable field
	var roleToGrant models.CircleRole
	var usedCount int
	var maxUses *int
	var expiresAt *time.Time

	err := h.DB.QueryRow(r.Context(),
		"SELECT id, circle_id, created_by_id, role_to_grant, used_count, max_uses, expires_at FROM invites WHERE code = $1",
		req.InviteCode).Scan(&inviteID, &circleID, &invitedByID, &roleToGrant, &usedCount, &maxUses, &expiresAt)

	if err != nil {
		http.Error(w, "Invalid invite code", http.StatusForbidden)
		return
	}

	if (maxUses != nil && usedCount >= *maxUses) || (expiresAt != nil && time.Now().After(*expiresAt)) {
		http.Error(w, "Invite code expired or used up", http.StatusForbidden)
		return
	}

	// 2. Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// 3. Create user and add to circle in a transaction
	tx, err := h.DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	var userID uuid.UUID
	err = tx.QueryRow(r.Context(),
		"INSERT INTO users (username, email, password_hash, invited_by_id) VALUES ($1, $2, $3, $4) RETURNING id",
		req.Username, req.Email, string(hash), invitedByID).Scan(&userID)
	if err != nil {
		http.Error(w, "Username or email already exists", http.StatusConflict)
		return
	}

	_, err = tx.Exec(r.Context(),
		"INSERT INTO circle_members (circle_id, user_id, invited_by_id, role) VALUES ($1, $2, $3, $4)",
		circleID, userID, invitedByID, roleToGrant)
	if err != nil {
		http.Error(w, "Failed to join circle", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(r.Context(), "UPDATE invites SET used_count = used_count + 1 WHERE id = $1", inviteID)
	if err != nil {
		http.Error(w, "Failed to update invite", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "Failed to commit registration", http.StatusInternalServerError)
		return
	}

	// 4. Create session
	h.createSession(w, r, userID)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var userID uuid.UUID
	var hash string
	var totpEnabled bool
	err := h.DB.QueryRow(r.Context(), "SELECT id, password_hash, totp_enabled FROM users WHERE username = $1", req.Username).Scan(&userID, &hash, &totpEnabled)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if totpEnabled {
		h.createMfaPendingSession(w, r, userID)
		return
	}

	h.createSession(w, r, userID)
}

func (h *Handler) createMfaPendingSession(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	token := generateToken()
	expiresAt := time.Now().Add(10 * time.Minute) // Short lived

	_, err := h.DB.Exec(r.Context(), "INSERT INTO sessions (token, user_id, expires_at, mfa_pending) VALUES ($1, $2, $3, TRUE)", token, userID, expiresAt)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "mfa_required"})
}

func (h *Handler) VerifyLoginTOTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Missing session", http.StatusUnauthorized)
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var userID uuid.UUID
	var expiresAt time.Time
	var mfaPending bool
	err = h.DB.QueryRow(r.Context(), "SELECT user_id, expires_at, mfa_pending FROM sessions WHERE token = $1", cookie.Value).Scan(&userID, &expiresAt, &mfaPending)
	if err != nil || !mfaPending || time.Now().After(expiresAt) {
		http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
		return
	}

	var secret string
	err = h.DB.QueryRow(r.Context(), "SELECT totp_secret FROM users WHERE id = $1", userID).Scan(&secret)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	if !totp.Validate(req.Code, secret) {
		http.Error(w, "Invalid MFA code", http.StatusUnauthorized)
		return
	}

	// Upgrade session
	newExpiresAt := time.Now().Add(24 * 7 * time.Hour)
	_, err = h.DB.Exec(r.Context(), "UPDATE sessions SET mfa_pending = FALSE, expires_at = $1 WHERE token = $2", newExpiresAt, cookie.Value)
	if err != nil {
		http.Error(w, "Failed to update session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    cookie.Value,
		Expires:  newExpiresAt,
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) SetupTOTP(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	var username string
	err := h.DB.QueryRow(r.Context(), "SELECT username FROM users WHERE id = $1", userID).Scan(&username)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Bulletin",
		AccountName: username,
	})
	if err != nil {
		http.Error(w, "Failed to generate TOTP key", http.StatusInternalServerError)
		return
	}

	// Store secret temporarily but don't enable MFA yet
	_, err = h.DB.Exec(r.Context(), "UPDATE users SET totp_secret = $1, totp_enabled = FALSE WHERE id = $2", key.Secret(), userID)
	if err != nil {
		http.Error(w, "Failed to store secret", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"secret": key.Secret(),
		"url":    key.URL(),
	})
}

func (h *Handler) EnableTOTP(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var secret string
	err := h.DB.QueryRow(r.Context(), "SELECT totp_secret FROM users WHERE id = $1", userID).Scan(&secret)
	if err != nil || secret == "" {
		http.Error(w, "TOTP not set up", http.StatusBadRequest)
		return
	}

	if !totp.Validate(req.Code, secret) {
		http.Error(w, "Invalid code", http.StatusUnauthorized)
		return
	}

	_, err = h.DB.Exec(r.Context(), "UPDATE users SET totp_enabled = TRUE WHERE id = $1", userID)
	if err != nil {
		http.Error(w, "Failed to enable TOTP", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DisableTOTP(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var hash string
	err := h.DB.QueryRow(r.Context(), "SELECT password_hash FROM users WHERE id = $1", userID).Scan(&hash)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	_, err = h.DB.Exec(r.Context(), "UPDATE users SET totp_enabled = FALSE, totp_secret = NULL WHERE id = $1", userID)
	if err != nil {
		http.Error(w, "Failed to disable TOTP", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	token := generateToken()
	expiresAt := time.Now().Add(24 * 7 * time.Hour) // 1 week

	_, err := h.DB.Exec(r.Context(), "INSERT INTO sessions (token, user_id, expires_at) VALUES ($1, $2, $3)", token, userID, expiresAt)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false, // Set to true in production
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	_, _ = h.DB.Exec(r.Context(), "DELETE FROM sessions WHERE token = $1", cookie.Value)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var userID uuid.UUID
	err := h.DB.QueryRow(r.Context(), "SELECT id FROM users WHERE email = $1", req.Email).Scan(&userID)
	if err != nil {
		// We don't want to leak if an email exists or not
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "If an account with that email exists, a reset link has been sent."})
		return
	}

	token := generateToken()
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err = h.DB.Exec(r.Context(),
		"INSERT INTO password_reset_tokens (token, user_id, expires_at) VALUES ($1, $2, $3)",
		token, userID, expiresAt)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send email
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)
	subject := "Password Reset Request"
	body := fmt.Sprintf("You requested a password reset. Click the link below to set a new password:\n\n%s\n\nThis link will expire in 1 hour.", resetURL)

	if h.Mailer != nil {
		err = h.Mailer.Send(req.Email, subject, body)
		if err != nil {
			log.Printf("Failed to send email: %v\n", err)
			http.Error(w, "Failed to send reset email", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "If an account with that email exists, a reset link has been sent."})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var userID uuid.UUID
	var expiresAt time.Time
	err := h.DB.QueryRow(r.Context(),
		"SELECT user_id, expires_at FROM password_reset_tokens WHERE token = $1",
		req.Token).Scan(&userID, &expiresAt)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusForbidden)
		return
	}

	if time.Now().After(expiresAt) {
		http.Error(w, "Token expired", http.StatusForbidden)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	tx, err := h.DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	_, err = tx.Exec(r.Context(), "UPDATE users SET password_hash = $1 WHERE id = $2", string(hash), userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(r.Context(), "DELETE FROM password_reset_tokens WHERE token = $1", req.Token)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Optionally invalidate all sessions for this user
	_, err = tx.Exec(r.Context(), "DELETE FROM sessions WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user models.User
	err := h.DB.QueryRow(r.Context(), "SELECT id, username, email, is_email_verified, totp_enabled, invited_by_id, created_at FROM users WHERE id = $1", userID).Scan(&user.ID, &user.Username, &user.Email, &user.IsEmailVerified, &user.TotpEnabled, &user.InvitedByID, &user.CreatedAt)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	var req struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		OldPassword string `json:"old_password"`
		Password    string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch current user data
	var currentUsername string
	var currentEmail *string
	var currentPasswordHash string
	err := h.DB.QueryRow(r.Context(), "SELECT username, email, password_hash FROM users WHERE id = $1", userID).Scan(&currentUsername, &currentEmail, &currentPasswordHash)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	tx, err := h.DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	if req.Username != "" && req.Username != currentUsername {
		_, err := tx.Exec(r.Context(), "UPDATE users SET username = $1 WHERE id = $2", req.Username, userID)
		if err != nil {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
	}

	// Handle email change
	if req.Email != "" && (currentEmail == nil || req.Email != *currentEmail) {
		// Check if email already exists
		var exists bool
		_ = tx.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)", req.Email, userID).Scan(&exists)
		if exists {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}

		// Update email and set is_email_verified to false
		_, err = tx.Exec(r.Context(), "UPDATE users SET email = $1, is_email_verified = FALSE WHERE id = $2", req.Email, userID)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		// We could automatically trigger verification here if we want
	}

	// Handle password change
	if req.Password != "" {
		if req.OldPassword == "" {
			http.Error(w, "Current password is required to set a new one", http.StatusBadRequest)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(currentPasswordHash), []byte(req.OldPassword)); err != nil {
			http.Error(w, "Incorrect current password", http.StatusUnauthorized)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		_, err = tx.Exec(r.Context(), "UPDATE users SET password_hash = $1 WHERE id = $2", string(hash), userID)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) RequestEmailVerification(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	var userEmail *string
	err := h.DB.QueryRow(r.Context(), "SELECT email FROM users WHERE id = $1", userID).Scan(&userEmail)
	if err != nil || userEmail == nil || *userEmail == "" {
		http.Error(w, "No email set for this user", http.StatusBadRequest)
		return
	}

	token := generateToken()
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = h.DB.Exec(r.Context(),
		"INSERT INTO email_verification_tokens (token, user_id, new_email, expires_at) VALUES ($1, $2, $3, $4)",
		token, userID, *userEmail, expiresAt)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send email
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, token)
	subject := "Verify your email address"
	body := fmt.Sprintf("Please click the link below to verify your email address:\n\n%s\n\nThis link will expire in 24 hours.", verifyURL)

	if h.Mailer != nil {
		err = h.Mailer.Send(*userEmail, subject, body)
		if err != nil {
			log.Printf("Failed to send verification email: %v\n", err)
			http.Error(w, "Failed to send verification email", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Verification email sent."})
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var userID uuid.UUID
	var newEmail string
	var expiresAt time.Time
	err := h.DB.QueryRow(r.Context(),
		"SELECT user_id, new_email, expires_at FROM email_verification_tokens WHERE token = $1",
		req.Token).Scan(&userID, &newEmail, &expiresAt)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusForbidden)
		return
	}

	if time.Now().After(expiresAt) {
		http.Error(w, "Token expired", http.StatusForbidden)
		return
	}

	tx, err := h.DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	// Verify that the email in the token matches the user's current email (to prevent verifying an old change)
	var currentEmail *string
	_ = tx.QueryRow(r.Context(), "SELECT email FROM users WHERE id = $1", userID).Scan(&currentEmail)
	if currentEmail == nil || *currentEmail != newEmail {
		http.Error(w, "Email mismatch", http.StatusBadRequest)
		return
	}

	_, err = tx.Exec(r.Context(), "UPDATE users SET is_email_verified = TRUE WHERE id = $1", userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(r.Context(), "DELETE FROM email_verification_tokens WHERE token = $1", req.Token)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
