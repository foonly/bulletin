package posts

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
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

func (h *Handler) JoinCircle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InviteCode string `json:"invite_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Verify invite code
	var inviteID uuid.UUID
	var circleID uuid.UUID
	var invitedByID *uuid.UUID
	var roleToGrant models.CircleRole
	var usedCount int
	var maxUses *int
	var expiresAt *time.Time

	err := h.DB.QueryRow(r.Context(),
		"SELECT id, circle_id, created_by_id, role_to_grant, used_count, max_uses, expires_at FROM invites WHERE code = $1",
		req.InviteCode).Scan(&inviteID, &circleID, &invitedByID, &roleToGrant, &usedCount, &maxUses, &expiresAt)

	if err != nil {
		http.Error(w, "Invalid invite code", http.StatusNotFound)
		return
	}

	if (maxUses != nil && usedCount >= *maxUses) || (expiresAt != nil && time.Now().After(*expiresAt)) {
		http.Error(w, "Invite code expired or used up", http.StatusGone)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	// Check if already a member
	var exists bool
	err = h.DB.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM circle_members WHERE circle_id = $1 AND user_id = $2)", circleID, userID).Scan(&exists)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "You are already a member of this circle", http.StatusConflict)
		return
	}

	tx, err := h.DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

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
		http.Error(w, "Failed to commit joining", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": circleID})
}

func (h *Handler) GetInvite(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	var res struct {
		CircleName string `json:"circle_name"`
		Valid      bool   `json:"valid"`
	}

	err := h.DB.QueryRow(r.Context(),
		`SELECT c.name, (i.max_uses IS NULL OR i.used_count < i.max_uses) AND (i.expires_at IS NULL OR i.expires_at > NOW())
		 FROM invites i
		 JOIN circles c ON i.circle_id = c.id
		 WHERE i.code = $1`, code).Scan(&res.CircleName, &res.Valid)

	if err != nil {
		http.Error(w, "Invite not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(res)
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
		`WITH RECURSIVE post_tree AS (
			SELECT p.id, p.id as root_id, p.circle_id, p.author_id, p.created_at
			FROM posts p
			JOIN circle_members cm ON p.circle_id = cm.circle_id
			WHERE p.parent_id IS NULL AND cm.user_id = $1
			UNION ALL
			SELECT p.id, pt.root_id, p.circle_id, p.author_id, p.created_at
			FROM posts p
			JOIN post_tree pt ON p.parent_id = pt.id
		),
		circle_unread_posts AS (
			SELECT pt.circle_id, COUNT(*) as count
			FROM post_tree pt
			LEFT JOIN read_markers rm ON rm.entity_id = pt.root_id AND rm.user_id = $1
			WHERE pt.author_id != $1
			AND pt.created_at > COALESCE(rm.last_read_at, '1970-01-01')
			GROUP BY pt.circle_id
		),
		circle_unread_chat AS (
			SELECT chat.circle_id, COUNT(*) as count
			FROM chat_messages chat
			JOIN circle_members cm ON chat.circle_id = cm.circle_id
			LEFT JOIN read_markers rm ON rm.entity_id = chat.circle_id AND rm.user_id = $1
			WHERE cm.user_id = $1
			AND chat.user_id != $1
			AND chat.created_at > COALESCE(rm.last_read_at, '1970-01-01')
			GROUP BY chat.circle_id
		),
		circle_member_counts AS (
			SELECT circle_id, COUNT(*) as count
			FROM circle_members
			GROUP BY circle_id
		),
		last_posts AS (
			-- Find the most recent post per circle
			SELECT DISTINCT ON (circle_id) circle_id, title, created_at
			FROM posts
			ORDER BY circle_id, created_at DESC
		)
		SELECT c.id, c.name, COALESCE(c.description, ''), c.owner_id, c.allow_freeform_tags,
		        c.invite_min_role, c.chat_retention_days, c.chat_retention_count, c.created_at, cm.role,
		        rm_chat.last_read_at,
		        COALESCE(cup.count, 0) + COALESCE(cuc.count, 0) as unread_count,
		        COALESCE(cuc.count, 0) as unread_chat_count,
		        COALESCE(cup.count, 0) as unread_post_count,
		        COALESCE(cmc.count, 0) as member_count,
		        lp.title as last_post_title,
		        lp.created_at as last_post_at
		 FROM circles c
		 JOIN circle_members cm ON c.id = cm.circle_id
		 LEFT JOIN read_markers rm_chat ON rm_chat.entity_id = c.id AND rm_chat.user_id = $1
		 LEFT JOIN circle_unread_posts cup ON cup.circle_id = c.id
		 LEFT JOIN circle_unread_chat cuc ON cuc.circle_id = c.id
		 LEFT JOIN circle_member_counts cmc ON cmc.circle_id = c.id
		 LEFT JOIN last_posts lp ON lp.circle_id = c.id
		 WHERE cm.user_id = $1`, userID)
	if err != nil {
		log.Printf("ListCircles query error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type circleWithRole struct {
		models.Circle
		Role models.CircleRole `json:"role"`
	}

	var circles []circleWithRole = []circleWithRole{}
	for rows.Next() {
		var c circleWithRole
		err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.OwnerID, &c.AllowFreeformTags,
			&c.InviteMinRole, &c.ChatRetentionDays, &c.ChatRetentionCount, &c.CreatedAt, &c.Role,
			&c.LastReadAt, &c.UnreadCount, &c.UnreadChatCount, &c.UnreadPostCount, &c.MemberCount,
			&c.LastPostTitle, &c.LastPostAt)
		if err != nil {
			log.Printf("ListCircles scan error: %v\n", err)
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

	// Update read marker for the author so their own post doesn't show as unread
	var readEntityID uuid.UUID
	if req.ParentID == nil {
		readEntityID = postID
	} else {
		// Find the root post ID for this thread
		err = tx.QueryRow(r.Context(), `
			WITH RECURSIVE thread_root AS (
				SELECT id, parent_id FROM posts WHERE id = $1
				UNION ALL
				SELECT p.id, p.parent_id FROM posts p JOIN thread_root tr ON tr.parent_id = p.id
			)
			SELECT id FROM thread_root WHERE parent_id IS NULL`, req.ParentID).Scan(&readEntityID)
		if err != nil {
			// If we can't find the root for some reason, just skip read marker update
			log.Printf("Failed to find root for read marker: %v\n", err)
		}
	}

	if readEntityID != uuid.Nil {
		_, _ = tx.Exec(r.Context(),
			`INSERT INTO read_markers (user_id, entity_id, last_read_at)
			 VALUES ($1, $2, NOW())
			 ON CONFLICT (user_id, entity_id) DO UPDATE SET last_read_at = NOW()`, userID, readEntityID)
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) MembershipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		circleIDStr := chi.URLParam(r, "circleID")
		circleID, err := uuid.Parse(circleIDStr)
		if err != nil {
			http.Error(w, "Invalid circle ID", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value("user_id").(uuid.UUID)
		var exists bool
		err = h.DB.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM circle_members WHERE circle_id = $1 AND user_id = $2)", circleID, userID).Scan(&exists)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !exists {
			http.Error(w, "Circle not found or access denied", http.StatusNotFound)
			return
		}

		next.ServeHTTP(w, r)
	})
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
			cm.role,
			cm.invited_by_id,
			inv.username as inviter_username,
			EXISTS(SELECT 1 FROM circle_members cm2 WHERE cm2.user_id = cm.invited_by_id AND cm2.circle_id IN (SELECT circle_id FROM circle_members WHERE user_id = $1)) as has_connection
		 FROM circle_members cm
		 JOIN users u ON cm.user_id = u.id
		 LEFT JOIN users inv ON cm.invited_by_id = inv.id
		 WHERE cm.circle_id = $2`, viewerID, circleID)

	if err != nil {
		log.Printf("ListMembers query error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type memberResponse struct {
		ID        uuid.UUID         `json:"id"`
		Username  string            `json:"username"`
		Role      models.CircleRole `json:"role"`
		InvitedBy string            `json:"invited_by"`
	}

	var members []memberResponse = []memberResponse{}
	for rows.Next() {
		var m memberResponse
		var invitedByID *uuid.UUID
		var inviterUsername *string
		var hasConnection bool
		err := rows.Scan(&m.ID, &m.Username, &m.Role, &invitedByID, &inviterUsername, &hasConnection)
		if err != nil {
			log.Printf("ListMembers scan error: %v\n", err)
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
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	ReplyCount  int        `json:"reply_count"`
	UnreadCount int        `json:"unread_count"`
	IsDeleted   bool       `json:"is_deleted"`
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

	// Optimized query: One recursive CTE to find all descendants of all threads in the circle
	query := `
		WITH RECURSIVE descendants AS (
			-- Find direct children of root posts in this circle
			SELECT p.id, p.parent_id, p.id as root_id, p.author_id, p.created_at
			FROM posts p
			WHERE p.circle_id = $1 AND p.parent_id IS NULL

			UNION ALL

			-- Find their children recursively
			SELECT p.id, p.parent_id, d.root_id, p.author_id, p.created_at
			FROM posts p
			JOIN descendants d ON p.parent_id = d.id
		),
		stats AS (
			-- Aggregate stats per root post
			SELECT
				root_id,
				COUNT(*) - 1 as total_replies, -- subtract 1 to exclude the root post itself
				MAX(created_at) as last_reply_at
			FROM descendants
			GROUP BY root_id
		),
		unread AS (
			-- Calculate unread counts per root post
			SELECT
				d.root_id,
				COUNT(*) as count
			FROM descendants d
			LEFT JOIN read_markers rm ON rm.entity_id = d.root_id AND rm.user_id = $2
			WHERE d.author_id != $2
			AND d.created_at > COALESCE(rm.last_read_at, '1970-01-01')
			GROUP BY d.root_id
		)
		SELECT
			p.id, p.author_id, u.username, COALESCE(p.title, ''), p.content, p.created_at, p.updated_at, p.is_deleted,
			COALESCE(s.total_replies, 0) as total_replies,
			s.last_reply_at,
			COALESCE(un.count, 0) as unread_count,
			COALESCE((SELECT string_agg(t.name, ',') FROM post_tags pt JOIN tags t ON pt.tag_id = t.id WHERE pt.post_id = p.id), '') as tags
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN stats s ON p.id = s.root_id
		LEFT JOIN unread un ON p.id = un.root_id
		WHERE p.circle_id = $1 AND p.parent_id IS NULL
	`

	args := []interface{}{circleID, userID}

	if tagFilter != "" {
		query += ` AND p.id IN (SELECT post_id FROM post_tags pt JOIN tags t ON pt.tag_id = t.id WHERE t.name = $3)`
		args = append(args, tagFilter)
	}

	query += ` ORDER BY COALESCE(s.last_reply_at, p.created_at) DESC`

	rows, err := h.DB.Query(r.Context(), query, args...)
	if err != nil {
		log.Printf("ListThreads query error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var threads []threadResponse = []threadResponse{}
	for rows.Next() {
		var t threadResponse
		var tagStr string
		err := rows.Scan(&t.ID, &t.AuthorID, &t.AuthorName, &t.Title, &t.Content, &t.CreatedAt, &t.UpdatedAt, &t.IsDeleted, &t.ReplyCount, &t.LastReplyAt, &t.UnreadCount, &tagStr)
		if err != nil {
			log.Printf("ListThreads scan error: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if tagStr != "" {
			t.Tags = strings.Split(tagStr, ",")
		} else {
			t.Tags = []string{}
		}
		threads = append(threads, t)
	}

	json.NewEncoder(w).Encode(threads)
}

func (h *Handler) GetThread(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "postID")
	postID, _ := uuid.Parse(postIDStr)
	userID := r.Context().Value("user_id").(uuid.UUID)

	// Fetch the entire thread tree using a recursive CTE
	rows, err := h.DB.Query(r.Context(),
		`WITH RECURSIVE thread_tree AS (
			-- Base case: the root post
			SELECT id, author_id, parent_id, title, content, created_at, updated_at, is_deleted, 0 as depth
			FROM posts
			WHERE id = $1

			UNION ALL

			-- Recursive step: find all children
			SELECT p.id, p.author_id, p.parent_id, p.title, p.content, p.created_at, p.updated_at, p.is_deleted, tt.depth + 1
			FROM posts p
			JOIN thread_tree tt ON p.parent_id = tt.id
		)
		SELECT tt.id, tt.author_id, u.username, tt.parent_id, COALESCE(tt.title, ''), tt.content, tt.created_at, tt.updated_at, tt.depth,
		       rm.last_read_at, tt.is_deleted
		FROM thread_tree tt
		JOIN users u ON tt.author_id = u.id
		LEFT JOIN read_markers rm ON rm.entity_id = $1 AND rm.user_id = $2
		ORDER BY tt.created_at ASC`, postID, userID)

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
		Title      string     `json:"title"`
		Content    string     `json:"content"`
		CreatedAt  time.Time  `json:"created_at"`
		UpdatedAt  *time.Time `json:"updated_at"`
		Depth      int        `json:"depth"`
		LastReadAt *time.Time `json:"last_read_at"`
		IsDeleted  bool       `json:"is_deleted"`
	}

	var posts []postNode = []postNode{}
	for rows.Next() {
		var p postNode
		err := rows.Scan(&p.ID, &p.AuthorID, &p.AuthorName, &p.ParentID, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt, &p.Depth, &p.LastReadAt, &p.IsDeleted)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		posts = append(posts, p)
	}

	json.NewEncoder(w).Encode(posts)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, _ := uuid.Parse(circleIDStr)
	query := r.URL.Query().Get("q")

	if query == "" {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	// Search for posts (threads or replies) matching the query
	// We return them with their thread root ID so we can link to the thread
	rows, err := h.DB.Query(r.Context(),
		`WITH RECURSIVE post_roots AS (
			SELECT id, id as root_id, parent_id
			FROM posts
			WHERE circle_id = $1 AND parent_id IS NULL
			UNION ALL
			SELECT p.id, pr.root_id, p.parent_id
			FROM posts p
			JOIN post_roots pr ON p.parent_id = pr.id
		)
		SELECT
			p.id, p.author_id, u.username, p.parent_id, pr.root_id,
			COALESCE(p.title, ''), p.content, p.created_at, p.updated_at, p.is_deleted
		FROM posts p
		JOIN users u ON p.author_id = u.id
		JOIN post_roots pr ON p.id = pr.id
		WHERE p.circle_id = $1
		  AND p.is_deleted = FALSE
		  AND (p.title ILIKE $2 OR p.content ILIKE $2)
		ORDER BY p.created_at DESC
		LIMIT 50`, circleID, "%"+query+"%")

	if err != nil {
		log.Printf("Search error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type searchResult struct {
		ID         uuid.UUID  `json:"id"`
		AuthorID   uuid.UUID  `json:"author_id"`
		AuthorName string     `json:"author_name"`
		ParentID   *uuid.UUID `json:"parent_id"`
		RootID     uuid.UUID  `json:"root_id"`
		Title      string     `json:"title"`
		Content    string     `json:"content"`
		CreatedAt  time.Time  `json:"created_at"`
		UpdatedAt  *time.Time `json:"updated_at"`
		IsDeleted  bool       `json:"is_deleted"`
	}

	var results []searchResult = []searchResult{}
	for rows.Next() {
		var res searchResult
		err := rows.Scan(&res.ID, &res.AuthorID, &res.AuthorName, &res.ParentID, &res.RootID, &res.Title, &res.Content, &res.CreatedAt, &res.UpdatedAt, &res.IsDeleted)
		if err != nil {
			log.Printf("Search scan error: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, res)
	}

	json.NewEncoder(w).Encode(results)
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, _ := uuid.Parse(circleIDStr)
	postIDStr := chi.URLParam(r, "postID")
	postID, _ := uuid.Parse(postIDStr)
	userID := r.Context().Value("user_id").(uuid.UUID)

	// Check if user is author OR admin/mod of the circle
	var authorID uuid.UUID
	var parentID *uuid.UUID
	err := h.DB.QueryRow(r.Context(), "SELECT author_id, parent_id FROM posts WHERE id = $1 AND circle_id = $2", postID, circleID).Scan(&authorID, &parentID)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	var userRole string
	err = h.DB.QueryRow(r.Context(), "SELECT role FROM circle_members WHERE circle_id = $1 AND user_id = $2", circleID, userID).Scan(&userRole)
	if err != nil {
		http.Error(w, "User not in circle", http.StatusForbidden)
		return
	}

	canDelete := authorID == userID || userRole == "admin" || userRole == "mod"
	if !canDelete {
		http.Error(w, "Not authorized to delete this post", http.StatusForbidden)
		return
	}

	if parentID == nil {
		// Hard delete thread
		_, err = h.DB.Exec(r.Context(), "DELETE FROM posts WHERE id = $1", postID)
	} else {
		// Soft delete reply
		_, err = h.DB.Exec(r.Context(), "UPDATE posts SET content = 'Reply has been deleted', is_deleted = TRUE WHERE id = $1", postID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "postID")
	postID, _ := uuid.Parse(postIDStr)

	userID := r.Context().Value("user_id").(uuid.UUID)

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verify authorship or admin role
	var authorID uuid.UUID
	var circleID uuid.UUID
	var isDeleted bool
	err := h.DB.QueryRow(r.Context(), "SELECT author_id, circle_id, is_deleted FROM posts WHERE id = $1", postID).Scan(&authorID, &circleID, &isDeleted)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	if isDeleted {
		http.Error(w, "Cannot edit a deleted post", http.StatusBadRequest)
		return
	}

	isAdmin, _ := h.checkRole(r.Context(), circleID, userID, models.RoleAdmin)

	if authorID != userID && !isAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_, err = h.DB.Exec(r.Context(), "UPDATE posts SET content = $1, updated_at = NOW() WHERE id = $2", req.Content, postID)
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

	if req.Name == "" {
		http.Error(w, "Circle name cannot be empty", http.StatusBadRequest)
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
		RoleToGrant  models.CircleRole `json:"role_to_grant"`
		MaxUses      *int              `json:"max_uses"`
		ExpiresInHrs *int              `json:"expires_in_hrs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate a random user-friendly code
	code := generateInviteCode()

	var expiresAt *time.Time
	if req.ExpiresInHrs != nil {
		t := time.Now().Add(time.Duration(*req.ExpiresInHrs) * time.Hour)
		expiresAt = &t
	}

	_, err = h.DB.Exec(r.Context(),
		`INSERT INTO invites (code, circle_id, created_by_id, role_to_grant, max_uses, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		code, circleID, userID, req.RoleToGrant, req.MaxUses, expiresAt)

	if err != nil {
		http.Error(w, "Failed to create invite", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"code": code})
}

func (h *Handler) ListInvites(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, _ := uuid.Parse(circleIDStr)

	rows, err := h.DB.Query(r.Context(),
		`SELECT i.id, i.code, i.role_to_grant, i.max_uses, i.used_count, i.expires_at, i.created_at, u.username
		 FROM invites i
		 JOIN users u ON i.created_by_id = u.id
		 WHERE i.circle_id = $1
		 ORDER BY i.created_at DESC`, circleID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type inviteResponse struct {
		ID          uuid.UUID         `json:"id"`
		Code        string            `json:"code"`
		RoleToGrant models.CircleRole `json:"role_to_grant"`
		MaxUses     *int              `json:"max_uses"`
		UsedCount   int               `json:"used_count"`
		ExpiresAt   *time.Time        `json:"expires_at"`
		CreatedAt   time.Time         `json:"created_at"`
		CreatedBy   string            `json:"created_by"`
	}

	var invites []inviteResponse = []inviteResponse{}
	for rows.Next() {
		var i inviteResponse
		err := rows.Scan(&i.ID, &i.Code, &i.RoleToGrant, &i.MaxUses, &i.UsedCount, &i.ExpiresAt, &i.CreatedAt, &i.CreatedBy)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		invites = append(invites, i)
	}

	json.NewEncoder(w).Encode(invites)
}

func (h *Handler) DeleteInvite(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, _ := uuid.Parse(circleIDStr)
	inviteIDStr := chi.URLParam(r, "inviteID")
	inviteID, _ := uuid.Parse(inviteIDStr)

	userID := r.Context().Value("user_id").(uuid.UUID)
	allowed, err := h.checkRole(r.Context(), circleID, userID, models.RoleMod) // Mods can revoke invites
	if err != nil || !allowed {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_, err = h.DB.Exec(r.Context(), "DELETE FROM invites WHERE id = $1 AND circle_id = $2", inviteID, circleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func generateInviteCode() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
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
	userID := r.Context().Value("user_id").(uuid.UUID)

	rows, err := h.DB.Query(r.Context(),
		`WITH RECURSIVE post_tree AS (
			SELECT id, id as root_id, author_id, created_at
			FROM posts
			WHERE circle_id = $1 AND parent_id IS NULL
			UNION ALL
			SELECT p.id, pt.root_id, p.author_id, p.created_at
			FROM posts p
			JOIN post_tree pt ON p.parent_id = pt.id
		),
		unread_posts AS (
			SELECT pt.id, pt.root_id
			FROM post_tree pt
			LEFT JOIN read_markers rm ON rm.entity_id = pt.root_id AND rm.user_id = $2
			WHERE pt.author_id != $2
			AND pt.created_at > COALESCE(rm.last_read_at, '1970-01-01')
		)
		SELECT
			t.id, t.name, t.is_pinned,
			COUNT(DISTINCT pt.post_id) as use_count,
			COUNT(DISTINCT up.id) as unread_count
		FROM tags t
		LEFT JOIN post_tags pt ON t.id = pt.tag_id
		LEFT JOIN unread_posts up ON up.root_id = pt.post_id
		WHERE t.circle_id = $1
		GROUP BY t.id, t.name, t.is_pinned
		ORDER BY t.is_pinned DESC, use_count DESC, t.name ASC`, circleID, userID)

	if err != nil {
		log.Printf("ListTags error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type tagResponse struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		IsPinned    bool      `json:"is_pinned"`
		UseCount    int       `json:"use_count"`
		UnreadCount int       `json:"unread_count"`
	}

	var tags []tagResponse = []tagResponse{}
	for rows.Next() {
		var t tagResponse
		err := rows.Scan(&t.ID, &t.Name, &t.IsPinned, &t.UseCount, &t.UnreadCount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tags = append(tags, t)
	}

	json.NewEncoder(w).Encode(tags)
}

func (h *Handler) CreateTag(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, _ := uuid.Parse(circleIDStr)

	userID := r.Context().Value("user_id").(uuid.UUID)
	allowed, err := h.checkRole(r.Context(), circleID, userID, models.RoleMod) // Mods can create tags
	if err != nil || !allowed {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(strings.ToLower(req.Name))
	if name == "" {
		http.Error(w, "Tag name is required", http.StatusBadRequest)
		return
	}

	_, err = h.DB.Exec(r.Context(),
		"INSERT INTO tags (circle_id, name) VALUES ($1, $2) ON CONFLICT (circle_id, name) DO NOTHING",
		circleID, name)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
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
