# API Reference

All API requests should be prefixed with `/api`. Authentication is handled via session cookies.

## Authentication

### `POST /auth/register`
Register a new user using an invite code.
- **Body**: `{ "username": "...", "password": "...", "invite_code": "..." }`

### `POST /auth/login`
- **Body**: `{ "username": "...", "password": "..." }`

### `POST /auth/logout`
Clears the session cookie and deletes the session from the database.

### `GET /auth/me`
Returns the current user's profile.

## Circles

### `GET /circles`
Returns a list of all circles the authenticated user is a member of.

### `GET /circles/{circleID}/members`
Returns a list of members in the circle, including "Invite Chain" information.
- **Privacy Logic**: The `invited_by` field shows the inviter's name ONLY if the viewer shares a circle with that inviter.

## Posts

### `GET /circles/{circleID}/posts`
Returns all posts in a circle.

### `POST /circles/{circleID}/posts`
Create a new post or reply.
- **Body**: `{ "title": "...", "content": "...", "parent_id": "UUID?", "tags": ["..."] }`

## Chat

### `GET /circles/{circleID}/chat/history`
Returns recent chat history based on the circle's retention policy.

### `GET /circles/{circleID}/chat/ws`
WebSocket endpoint for real-time chat.
- **Message Format**: `{ "content": "..." }`
