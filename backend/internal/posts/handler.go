package posts

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/username/bulletin/backend/internal/models"
)

type Handler struct {
	DB *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{DB: db}
}

func (h *Handler) ListCircles(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	rows, err := h.DB.Query(r.Context(),
		`SELECT c.id, c.name, c.description, c.owner_id, c.allow_freeform_tags,
		        c.invite_min_role, c.chat_retention_days, c.chat_retention_count, c.created_at, cm.role
		 FROM circles c
		 JOIN circle_members cm ON c.id = cm.circle_id
		 WHERE cm.user_id = $1`, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type circleWithRole struct {
		models.Circle
		Role models.CircleRole `json:"role"`
	}

	var circles []circleWithRole
	for rows.Next() {
		var c circleWithRole
		err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.OwnerID, &c.AllowFreeformTags,
			&c.InviteMinRole, &c.ChatRetentionDays, &c.ChatRetentionCount, &c.CreatedAt, &c.Role)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		circles = append(circles, c)
	}

	json.NewEncoder(w).Encode(circles)
}

type postResponse struct {
	ID         uuid.UUID  `json:"id"`
	AuthorID   uuid.UUID  `json:"author_id"`
	AuthorName string     `json:"author_name"`
	ParentID   *uuid.UUID `json:"parent_id"`
	Title      string     `json:"title,omitempty"`
	Content    string     `json:"content"`
	CreatedAt  string     `json:"created_at"`
	Tags       []string   `json:"tags,omitempty"`
}

func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, err := uuid.Parse(circleIDStr)
	if err != nil {
		http.Error(w, "Invalid circle ID", http.StatusBadRequest)
		return
	}

	// Verify membership
	userID := r.Context().Value("user_id").(uuid.UUID)
	var exists bool
	err = h.DB.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM circle_members WHERE circle_id = $1 AND user_id = $2)", circleID, userID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	rows, err := h.DB.Query(r.Context(),
		`SELECT p.id, p.author_id, u.username, p.parent_id, p.title, p.content, p.created_at
		 FROM posts p
		 JOIN users u ON p.author_id = u.id
		 WHERE p.circle_id = $1
		 ORDER BY p.created_at ASC`, circleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []postResponse
	for rows.Next() {
		var p postResponse
		var createdAt time.Time
		err := rows.Scan(&p.ID, &p.AuthorID, &p.AuthorName, &p.ParentID, &p.Title, &p.Content, &createdAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		p.CreatedAt = createdAt.Format(time.RFC3339)
		posts = append(posts, p)
	}

	json.NewEncoder(w).Encode(posts)
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, err := uuid.Parse(circleIDStr)
	if err != nil {
		http.Error(w, "Invalid circle ID", http.StatusBadRequest)
		return
	}

	var req struct {
		ParentID *uuid.UUID `json:"parent_id"`
		Title    string     `json:"title"`
		Content  string     `json:"content"`
		Tags     []string   `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	tx, err := h.DB.Begin(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	var postID uuid.UUID
	err = tx.QueryRow(r.Context(),
		"INSERT INTO posts (circle_id, author_id, parent_id, title, content) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		circleID, userID, req.ParentID, req.Title, req.Content).Scan(&postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle tags
	for _, tagName := range req.Tags {
		var tagID uuid.UUID
		err = tx.QueryRow(r.Context(),
			"INSERT INTO tags (circle_id, name) VALUES ($1, $2) ON CONFLICT (circle_id, name) DO UPDATE SET name = EXCLUDED.name RETURNING id",
			circleID, tagName).Scan(&tagID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(r.Context(), "INSERT INTO post_tags (post_id, tag_id) VALUES ($1, $2)", postID, tagID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, err := uuid.Parse(circleIDStr)
	if err != nil {
		http.Error(w, "Invalid circle ID", http.StatusBadRequest)
		return
	}

	viewerID := r.Context().Value("user_id").(uuid.UUID)

	rows, err := h.DB.Query(r.Context(),
		`SELECT
			u.id,
			u.username,
			cm.invited_by_id,
			inv.username as inviter_username,
			EXISTS(SELECT 1 FROM circle_members cm2 WHERE cm2.user_id = cm.invited_by_id AND cm2.circle_id IN (SELECT circle_id FROM circle_members WHERE user_id = $1)) as has_connection
		 FROM circle_members cm
		 JOIN users u ON cm.user_id = u.id
		 LEFT JOIN users inv ON cm.invited_by_id = inv.id
		 WHERE cm.circle_id = $2`, viewerID, circleID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type memberResponse struct {
		ID        uuid.UUID `json:"id"`
		Username  string    `json:"username"`
		InvitedBy string    `json:"invited_by"`
	}

	var members []memberResponse
	for rows.Next() {
		var m memberResponse
		var invitedByID *uuid.UUID
		var inviterUsername *string
		var hasConnection bool
		err := rows.Scan(&m.ID, &m.Username, &invitedByID, &inviterUsername, &hasConnection)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if invitedByID == nil {
			m.InvitedBy = "System"
		} else if hasConnection || *invitedByID == viewerID {
			m.InvitedBy = *inviterUsername
		} else {
			m.InvitedBy = "Unknown"
		}
		members = append(members, m)
	}

	json.NewEncoder(w).Encode(members)
}

func (h *Handler) UpdateCircle(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, err := uuid.Parse(circleIDStr)
	if err != nil {
		http.Error(w, "Invalid circle ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)
	allowed, err := h.checkRole(r.Context(), circleID, userID, models.RoleAdmin)
	if err != nil || !allowed {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		Name               string            `json:"name"`
		Description        string            `json:"description"`
		AllowFreeformTags  bool              `json:"allow_freeform_tags"`
		InviteMinRole      models.CircleRole `json:"invite_min_role"`
		ChatRetentionDays  int               `json:"chat_retention_days"`
		ChatRetentionCount int               `json:"chat_retention_count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = h.DB.Exec(r.Context(),
		`UPDATE circles
		 SET name = $1, description = $2, allow_freeform_tags = $3, invite_min_role = $4,
		     chat_retention_days = $5, chat_retention_count = $6
		 WHERE id = $7`,
		req.Name, req.Description, req.AllowFreeformTags, req.InviteMinRole,
		req.ChatRetentionDays, req.ChatRetentionCount, circleID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, err := uuid.Parse(circleIDStr)
	if err != nil {
		http.Error(w, "Invalid circle ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	var minRole models.CircleRole
	err = h.DB.QueryRow(r.Context(), "SELECT invite_min_role FROM circles WHERE id = $1", circleID).Scan(&minRole)
	if err != nil {
		http.Error(w, "Circle not found", http.StatusNotFound)
		return
	}

	allowed, err := h.checkRole(r.Context(), circleID, userID, minRole)
	if err != nil || !allowed {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		Code         string            `json:"code"`
		RoleToGrant  models.CircleRole `json:"role_to_grant"`
		MaxUses      *int              `json:"max_uses"`
		ExpiresInHrs *int              `json:"expires_in_hrs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var expiresAt *time.Time
	if req.ExpiresInHrs != nil {
		t := time.Now().Add(time.Duration(*req.ExpiresInHrs) * time.Hour)
		expiresAt = &t
	}

	_, err = h.DB.Exec(r.Context(),
		`INSERT INTO invites (code, circle_id, created_by_id, role_to_grant, max_uses, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		req.Code, circleID, userID, req.RoleToGrant, req.MaxUses, expiresAt)

	if err != nil {
		http.Error(w, "Invite code already exists or server error", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, _ := uuid.Parse(circleIDStr)
	targetUserIDStr := chi.URLParam(r, "userID")
	targetUserID, _ := uuid.Parse(targetUserIDStr)

	userID := r.Context().Value("user_id").(uuid.UUID)
	allowed, err := h.checkRole(r.Context(), circleID, userID, models.RoleAdmin)
	if err != nil || !allowed {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		Role models.CircleRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = h.DB.Exec(r.Context(), "UPDATE circle_members SET role = $1 WHERE circle_id = $2 AND user_id = $3", req.Role, circleID, targetUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteMember(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, _ := uuid.Parse(circleIDStr)
	targetUserIDStr := chi.URLParam(r, "userID")
	targetUserID, _ := uuid.Parse(targetUserIDStr)

	userID := r.Context().Value("user_id").(uuid.UUID)
	allowed, err := h.checkRole(r.Context(), circleID, userID, models.RoleAdmin)
	if err != nil || !allowed {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_, err = h.DB.Exec(r.Context(), "DELETE FROM circle_members WHERE circle_id = $1 AND user_id = $2", circleID, targetUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) checkRole(ctx context.Context, circleID, userID uuid.UUID, minRole models.CircleRole) (bool, error) {
	var role models.CircleRole
	err := h.DB.QueryRow(ctx, "SELECT role FROM circle_members WHERE circle_id = $1 AND user_id = $2", circleID, userID).Scan(&role)
	if err != nil {
		return false, err
	}

	roles := map[models.CircleRole]int{
		models.RoleGuest:    0,
		models.RoleStandard: 1,
		models.RoleMod:      2,
		models.RoleAdmin:    3,
	}

	return roles[role] >= roles[minRole], nil
}
