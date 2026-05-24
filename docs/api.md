# API Reference

All API requests are prefixed with `/api`. Authenticated endpoints require a valid session cookie.

## Authentication

### `POST /auth/register`

- **Body**: `{ "username": "...", "password": "...", "invite_code": "..." }`
- **Role**: Public. Creates user and joins them to the invite's target circle.

### `POST /auth/login`

- **Body**: `{ "username": "...", "password": "..." }`
- **Role**: Public. Establishes a session.

### `GET /auth/me`

- **Role**: Authenticated. Returns current user profile.

### `PUT /auth/me`

- **Body**: `{ "username": "?", "password": "?" }`
- **Role**: Authenticated. Updates username or password.

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
