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

func (h *Handler) CreateCircle(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Circle name is required", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	var circleID uuid.UUID
	err = tx.QueryRow(r.Context(),
		`INSERT INTO circles (name, description, owner_id)
		 VALUES ($1, $2, $3) RETURNING id`,
		req.Name, req.Description, userID).Scan(&circleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(r.Context(),
		`INSERT INTO circle_members (circle_id, user_id, role)
		 VALUES ($1, $2, $3)`,
		circleID, userID, models.RoleAdmin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": circleID})
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

	// For new threads (parent_id == nil), ensure at least one tag is provided
	if req.ParentID == nil && len(req.Tags) == 0 {
		http.Error(w, "At least one tag is required for new threads", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	// Check if circle allows freeform tags
	var allowFreeform bool
	err = h.DB.QueryRow(r.Context(), "SELECT allow_freeform_tags FROM circles WHERE id = $1", circleID).Scan(&allowFreeform)
	if err != nil {
		http.Error(w, "Circle not found", http.StatusNotFound)
		return
	}

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
		if allowFreeform {
			err = tx.QueryRow(r.Context(),
				"INSERT INTO tags (circle_id, name) VALUES ($1, $2) ON CONFLICT (circle_id, name) DO UPDATE SET name = EXCLUDED.name RETURNING id",
				circleID, tagName).Scan(&tagID)
		} else {
			err = tx.QueryRow(r.Context(),
				"SELECT id FROM tags WHERE circle_id = $1 AND name = $2",
				circleID, tagName).Scan(&tagID)
		}

		if err != nil {
			if !allowFreeform {
				http.Error(w, "Tag '"+tagName+"' does not exist and freeform tags are not allowed", http.StatusBadRequest)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
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

type threadResponse struct {
	ID          uuid.UUID  `json:"id"`
	AuthorID    uuid.UUID  `json:"author_id"`
	AuthorName  string     `json:"author_name"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	CreatedAt   time.Time  `json:"created_at"`
	ReplyCount  int        `json:"reply_count"`
	UnreadCount int        `json:"unread_count"`
	LastReplyAt *time.Time `json:"last_reply_at"`
	Tags        []string   `json:"tags"`
}

func (h *Handler) ListThreads(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, err := uuid.Parse(circleIDStr)
	if err != nil {
		http.Error(w, "Invalid circle ID", http.StatusBadRequest)
		return
	}

	tagFilter := r.URL.Query().Get("tag")

	userID := r.Context().Value("user_id").(uuid.UUID)

	query := `SELECT
			p.id, p.author_id, u.username, p.title, p.content, p.created_at,
			(
				WITH RECURSIVE descendants AS (
					SELECT id FROM posts WHERE parent_id = p.id
					UNION ALL
					SELECT p2.id FROM posts p2 JOIN descendants d ON p2.parent_id = d.id
				)
				SELECT COUNT(*) FROM descendants
			) as total_replies,
			(
				WITH RECURSIVE descendants AS (
					SELECT id, created_at FROM posts WHERE parent_id = p.id
					UNION ALL
					SELECT p2.id, p2.created_at FROM posts p2 JOIN descendants d ON p2.parent_id = d.id
				)
				SELECT MAX(created_at) FROM descendants
			) as last_reply_at,
			(
				WITH RECURSIVE descendants AS (
					SELECT id, created_at FROM posts WHERE parent_id = p.id
					UNION ALL
					SELECT p2.id, p2.created_at FROM posts p2 JOIN descendants d ON p2.parent_id = d.id
				)
				SELECT COUNT(*) FROM descendants WHERE created_at > COALESCE(rm.last_read_at, '1970-01-01')
			) as unread_count,
			COALESCE((SELECT array_agg(t.name) FROM post_tags pt JOIN tags t ON pt.tag_id = t.id WHERE pt.post_id = p.id), '{}') as tags
		 FROM posts p
		 JOIN users u ON p.author_id = u.id
		 LEFT JOIN read_markers rm ON rm.entity_id = p.id AND rm.user_id = $2
		 WHERE p.circle_id = $1 AND p.parent_id IS NULL`

	args := []interface{}{circleID, userID}

	if tagFilter != "" {
		query += ` AND p.id IN (SELECT post_id FROM post_tags pt JOIN tags t ON pt.tag_id = t.id WHERE t.name = $3)`
		args = append(args, tagFilter)
	}

	query += ` ORDER BY COALESCE((
				WITH RECURSIVE descendants AS (
					SELECT id, created_at FROM posts WHERE parent_id = p.id
					UNION ALL
					SELECT p2.id, p2.created_at FROM posts p2 JOIN descendants d ON p2.parent_id = d.id
				)
				SELECT MAX(created_at) FROM descendants
		 ), p.created_at) DESC`

	rows, err := h.DB.Query(r.Context(), query, args...)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var threads []threadResponse = []threadResponse{}
	for rows.Next() {
		var t threadResponse
		err := rows.Scan(&t.ID, &t.AuthorID, &t.AuthorName, &t.Title, &t.Content, &t.CreatedAt, &t.ReplyCount, &t.LastReplyAt, &t.UnreadCount, &t.Tags)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		threads = append(threads, t)
	}

	json.NewEncoder(w).Encode(threads)
}

func (h *Handler) GetThread(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "postID")
	postID, _ := uuid.Parse(postIDStr)

	// Fetch the entire thread tree using a recursive CTE
	rows, err := h.DB.Query(r.Context(),
		`WITH RECURSIVE thread_tree AS (
			-- Base case: the root post
			SELECT id, author_id, parent_id, title, content, created_at, 0 as depth
			FROM posts
			WHERE id = $1

			UNION ALL

			-- Recursive step: find all children
			SELECT p.id, p.author_id, p.parent_id, p.title, p.content, p.created_at, tt.depth + 1
			FROM posts p
			JOIN thread_tree tt ON p.parent_id = tt.id
		)
		SELECT tt.id, tt.author_id, u.username, tt.parent_id, COALESCE(tt.title, ''), tt.content, tt.created_at, tt.depth
		FROM thread_tree tt
		JOIN users u ON tt.author_id = u.id
		ORDER BY tt.created_at ASC`, postID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type postNode struct {
		ID         uuid.UUID  `json:"id"`
		AuthorID   uuid.UUID  `json:"author_id"`
		AuthorName string     `json:"author_name"`
		ParentID   *uuid.UUID `json:"parent_id"`
		Title      string     `json:"title,omitempty"`
		Content    string     `json:"content"`
		CreatedAt  time.Time  `json:"created_at"`
		Depth      int        `json:"depth"`
	}

	var allPosts []postNode
	for rows.Next() {
		var p postNode
		err := rows.Scan(&p.ID, &p.AuthorID, &p.AuthorName, &p.ParentID, &p.Title, &p.Content, &p.CreatedAt, &p.Depth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		allPosts = append(allPosts, p)
	}

	if len(allPosts) == 0 {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(allPosts)
}

func (h *Handler) UpdateReadMarker(w http.ResponseWriter, r *http.Request) {
	entityIDStr := chi.URLParam(r, "entityID")
	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	_, err = h.DB.Exec(r.Context(),
		`INSERT INTO read_markers (user_id, entity_id, last_read_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (user_id, entity_id) DO UPDATE SET last_read_at = NOW()`, userID, entityID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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

func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, _ := uuid.Parse(circleIDStr)

	rows, err := h.DB.Query(r.Context(),
		`SELECT t.id, t.name, t.is_pinned, COUNT(pt.post_id) as use_count
		 FROM tags t
		 LEFT JOIN post_tags pt ON t.id = pt.tag_id
		 WHERE t.circle_id = $1
		 GROUP BY t.id, t.name, t.is_pinned
		 ORDER BY t.is_pinned DESC, use_count DESC, t.name ASC`, circleID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type tagResponse struct {
		ID       uuid.UUID `json:"id"`
		Name     string    `json:"name"`
		IsPinned bool      `json:"is_pinned"`
		UseCount int       `json:"use_count"`
	}

	var tags []tagResponse = []tagResponse{}
	for rows.Next() {
		var t tagResponse
		err := rows.Scan(&t.ID, &t.Name, &t.IsPinned, &t.UseCount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tags = append(tags, t)
	}

	json.NewEncoder(w).Encode(tags)
}

func (h *Handler) PinTag(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, _ := uuid.Parse(circleIDStr)
	tagIDStr := chi.URLParam(r, "tagID")
	tagID, _ := uuid.Parse(tagIDStr)

	userID := r.Context().Value("user_id").(uuid.UUID)
	allowed, err := h.checkRole(r.Context(), circleID, userID, models.RoleAdmin)
	if err != nil || !allowed {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		IsPinned bool `json:"is_pinned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = h.DB.Exec(r.Context(), "UPDATE tags SET is_pinned = $1 WHERE id = $2 AND circle_id = $3", req.IsPinned, tagID, circleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
