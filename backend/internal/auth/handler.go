package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/username/bulletin/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	DB *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{DB: db}
}

type registerRequest struct {
	Username   string `json:"username"`
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
		"INSERT INTO users (username, password_hash, invited_by_id) VALUES ($1, $2, $3) RETURNING id",
		req.Username, string(hash), invitedByID).Scan(&userID)
	if err != nil {
		http.Error(w, "Username already exists", http.StatusConflict)
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
	err := h.DB.QueryRow(r.Context(), "SELECT id, password_hash FROM users WHERE username = $1", req.Username).Scan(&userID, &hash)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	h.createSession(w, r, userID)
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

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user models.User
	err := h.DB.QueryRow(r.Context(), "SELECT id, username, invited_by_id, created_at FROM users WHERE id = $1", userID).Scan(&user.ID, &user.Username, &user.InvitedByID, &user.CreatedAt)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Username != "" {
		_, err := h.DB.Exec(r.Context(), "UPDATE users SET username = $1 WHERE id = $2", req.Username, userID)
		if err != nil {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
	}

	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		_, err = h.DB.Exec(r.Context(), "UPDATE users SET password_hash = $1 WHERE id = $2", string(hash), userID)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
