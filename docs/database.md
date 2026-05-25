# Database Schema

Bulletin uses PostgreSQL 16 as its primary data store. The schema is designed for deep threading and efficient unread tracking.

## Core Tables

### `users`

Primary identity table.

- `id` (UUID): Primary key.
- `username` (TEXT): Unique username.
- `email` (TEXT): Unique email address.
- `password_hash` (TEXT): Bcrypt hashed password.
- `is_email_verified` (BOOL): Default FALSE.
- `totp_secret` (TEXT): Encrypted TOTP secret.
- `totp_enabled` (BOOL): Default FALSE.
- `invited_by_id` (UUID): FK to `users.id` (tracks the inviter).
- `created_at` (TIMESTAMPTZ).

### `sessions`

Database-backed session storage.

- `token` (TEXT): Primary key (random hex).
- `user_id` (UUID): FK to `users.id`.
- `mfa_pending` (BOOL): TRUE if user still needs to provide TOTP code.
- `expires_at` (TIMESTAMPTZ).

### `circles`

Group management.

- `id` (UUID): Primary key.
- `name` (TEXT): Required.
- `description` (TEXT): Optional.
- `owner_id` (UUID): FK to `users.id`.
- `allow_freeform_tags` (BOOL): Default TRUE.
- `invite_min_role` (ENUM): Role required to generate invites.
- `chat_retention_days` (INT): Default 14.
- `chat_retention_count` (INT): Default 50.

### `circle_members`

Many-to-many relationship between users and circles.

- `circle_id`, `user_id`: Composite primary key.
- `invited_by_id` (UUID): Who invited this user to this specific circle.
- `role` (ENUM): `guest`, `standard`, `mod`, `admin`.

### `invites`

Invite code tracking.

- `code` (TEXT): Unique 12-char hex string.
- `max_uses` (INT): Optional limit.
- `used_count` (INT).
- `expires_at` (TIMESTAMPTZ): Optional expiration.

## Conversations

### `posts`

Storage for both top-level threads and nested replies.

- `id` (UUID): Primary key.
- `parent_id` (UUID): FK to `posts.id` (NULL for top-level threads).
- `circle_id` (UUID): FK to `circles.id`.
- `author_id` (UUID): FK to `users.id`.
- `title` (TEXT): Only present for root posts.
- `content` (TEXT): Markdown supported content.
- `is_deleted` (BOOL): Default FALSE (soft-delete for replies).
- `updated_at` (TIMESTAMPTZ): Tracks last edit.

### `chat_messages`

High-frequency real-time messaging.

- `circle_id` (UUID): Indexed for fast history lookup.
- `user_id` (UUID): Sender.
- `content` (TEXT).

### `read_markers`

Tracks reading progress for unread indicators.

- `user_id`, `entity_id`: Primary key.
- `entity_id` can be a `post_id` (thread) or `circle_id` (chat).
- `last_read_at` (TIMESTAMPTZ).

## Taxonomy

### `tags`

Categorization per circle.

- `circle_id`, `name`: Unique constraint.
- `is_pinned` (BOOL): Pinned to the top of the navigation.

### `post_tags`

Links tags to root posts.

- `post_id`, `tag_id`: Composite primary key.

## Security & System

### `schema_migrations`

Tracks applied database migrations.

- `name` (TEXT): Primary key (filename).
- `applied_at` (TIMESTAMPTZ).

### `password_reset_tokens`

- `token` (TEXT): Primary key.
- `user_id` (UUID): FK to `users.id`.
- `expires_at` (TIMESTAMPTZ).

### `email_verification_tokens`

- `token` (TEXT): Primary key.
- `user_id` (UUID): FK to `users.id`.
- `new_email` (TEXT): Target email to verify.
- `expires_at` (TIMESTAMPTZ).
