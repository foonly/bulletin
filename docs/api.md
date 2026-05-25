# API Reference

All API requests are prefixed with `/api`. Authenticated endpoints require a valid session cookie.

## Authentication

### `POST /auth/register`

- **Body**: `{ "username": "...", "email": "...", "password": "...", "invite_code": "..." }`
- **Role**: Public. Creates user and joins them to the invite's target circle.

### `POST /auth/login`

- **Body**: `{ "username": "...", "password": "..." }`
- **Response**: `{ "status": "success" | "mfa_required" }`
- **Role**: Public. Establishes a session.

### `POST /auth/login-totp`

- **Body**: `{ "code": "6-digit-string" }`
- **Role**: MFA-Pending. Verifies TOTP code and upgrades session.

### `POST /auth/request-reset`

- **Body**: `{ "email": "..." }`
- **Role**: Public. Sends a password reset email.

### `POST /auth/reset-password`

- **Body**: `{ "token": "...", "password": "..." }`
- **Role**: Public. Sets a new password using a reset token.

### `POST /auth/request-verification`

- **Role**: Authenticated. Sends an email verification link.

### `POST /auth/verify-email`

- **Body**: `{ "token": "..." }`
- **Role**: Public. Marks user email as verified.

### `GET /auth/me`

- **Role**: Authenticated. Returns current user profile.

### `PUT /auth/me`

- **Body**: `{ "username": "?", "email": "?", "old_password": "?", "password": "?" }`
- **Role**: Authenticated. Updates user profile. Changing password requires `old_password`.

### `POST /auth/totp/setup`

- **Role**: Authenticated. Generates a new TOTP secret.

### `POST /auth/totp/enable`

- **Body**: `{ "code": "..." }`
- **Role**: Authenticated. Verifies code and enables MFA.

### `POST /auth/totp/disable`

- **Body**: `{ "password": "..." }`
- **Role**: Authenticated. Disables MFA.

## Circles (Global)

### `GET /circles`

- **Role**: Authenticated. Lists circles the user is a member of.

### `POST /circles`

- **Body**: `{ "name": "...", "description": "..." }`
- **Role**: Authenticated. Creates a new circle and grants creator `admin` role.

## Circle-Scoped (`/circles/{circleID}/*`)

_All endpoints below are protected by Membership Middleware._

### `PUT /`

- **Body**: Circle configuration (retention, invite rules, etc.).
- **Role**: Admin/Mod.

### `GET /threads`

- **Query**: `?tag=NAME` (optional).
- **Role**: Member. Returns top-level posts with unread counts and reply stats.

### `GET /threads/{postID}`

- **Role**: Member. Returns the full recursive tree of posts/replies for a thread.

### `PUT /threads/{postID}`

- **Body**: `{ "content": "..." }`
- **Role**: Author or Admin.

### `DELETE /threads/{postID}`

- **Role**: Author or Admin. Hard-deletes threads; soft-deletes replies.

### `GET /search`

- **Query**: `?q=query`
- **Role**: Member. Searches titles and content within the circle.

### `POST /posts`

- **Body**: `{ "title": "?", "content": "...", "parent_id": "UUID?", "tags": ["..."] }`
- **Role**: Member. Creates a new thread (requires tags) or a reply.

### `GET /members`

- **Role**: Member. Returns list of members with roles and invite origins.

### `PUT /members/{userID}`

- **Body**: `{ "role": "..." }`
- **Role**: Admin. Updates a user's role in the circle.

### `DELETE /members/{userID}`

- **Role**: Admin. Kicks a user from the circle.

### `GET /tags`

- **Role**: Member. Lists tags ordered by pinned status and popularity.

### `POST /tags`

- **Body**: `{ "name": "..." }`
- **Role**: Admin/Mod. Pre-creates a tag.

### `POST /tags/{tagID}/pin`

- **Body**: `{ "is_pinned": bool }`
- **Role**: Admin/Mod. Pins/unpins a tag.

### `GET /invites`

- **Role**: Admin/Mod. Lists active invite codes.

### `POST /invites`

- **Body**: `{ "role_to_grant": "...", "max_uses": int?, "expires_in_hrs": int? }`
- **Role**: Based on Circle Settings. Generates a secure random code.

### `DELETE /invites/{inviteID}`

- **Role**: Admin/Mod. Revokes an invite code.

### `POST /read/{entityID}`

- **Role**: Member. Marks a thread or chat as read for the user.

### `GET /chat/ws`

- **Role**: Member. WebSocket upgrade endpoint.
