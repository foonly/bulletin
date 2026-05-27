package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID  `json:"id"`
	Username        string     `json:"username"`
	Email           *string    `json:"email"`
	IsEmailVerified bool       `json:"is_email_verified"`
	TotpEnabled     bool       `json:"totp_enabled"`
	TotpSecret      *string    `json:"-"`
	PasswordHash    string     `json:"-"`
	InvitedByID     *uuid.UUID `json:"invited_by_id"`
	CreatedAt       time.Time  `json:"created_at"`
}

type Session struct {
	Token     string    `json:"token"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type CircleRole string

const (
	RoleGuest    CircleRole = "guest"
	RoleStandard CircleRole = "standard"
	RoleMod      CircleRole = "mod"
	RoleAdmin    CircleRole = "admin"
)

type Circle struct {
	ID                 uuid.UUID  `json:"id"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	OwnerID            *uuid.UUID `json:"owner_id"`
	AllowFreeformTags  bool       `json:"allow_freeform_tags"`
	InviteMinRole      CircleRole `json:"invite_min_role"`
	ChatRetentionDays  int        `json:"chat_retention_days"`
	ChatRetentionCount int        `json:"chat_retention_count"`
	Palette            string     `json:"palette"`
	CreatedAt          time.Time  `json:"created_at"`
	LastReadAt         *time.Time `json:"last_read_at,omitempty"`
	UnreadCount        int        `json:"unread_count"`
	UnreadChatCount    int        `json:"unread_chat_count"`
	UnreadPostCount    int        `json:"unread_post_count"`
	MemberCount        int        `json:"member_count"`
	LastPostTitle      *string    `json:"last_post_title"`
	LastPostAt         *time.Time `json:"last_post_at"`
	IsDeleted          bool       `json:"is_deleted"`
}
