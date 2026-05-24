# Bulletin Documentation

Welcome to the Bulletin technical documentation. Bulletin is a communications system based on social circles, featuring threaded posts and real-time chat.

## Table of Contents

1. [Architecture Overview](./architecture.md)
2. [Database Schema](./database.md)
3. [API Reference](./api.md)
4. [Development Guide](./development.md)
5. [Invite System & Permissions](./invites.md)

## Completed Features

- **Social Circles**: Create private groups with custom names, descriptions, and settings.
- **Threaded Conversations**: Deeply nested post/reply system with recursive rendering and smart unread tracking.
- **Real-time Chat**: Instant messaging with per-circle presence (online/offline indicators).
- **Advanced Tagging**: Organize threads with mandatory tags, support for pinning important categories, and filtered navigation.
- **Robust Invite System**: Automatic secure code generation with usage limits, expiration, and full "issued-by" audit trail.
- **Privacy & Security**: Membership-based middleware protection for all routes and view-based error states.
- **Read Tracking**: Automatic read markers for both threaded conversations and chat history.
- **Automated Retention**: Background worker to purge old chat history based on per-circle policies (e.g., 14 days or 50 messages).
- **User Management**: Session-based auth with username/password updates and administrative role management.
- **Nimble Dev DX**: Live-reloading for both Go (Air) and Vue (Vite), with infrastructure isolated in Docker.

## Technology Stack

- **Backend**: Go 1.22+ (chi, pgx, gorilla/websocket, air)
- **Frontend**: Vue 3 (Pinia, Vue Router, Tailwind CSS v4)
- **Database**: PostgreSQL 16 (Recursive CTEs, Advanced JSON/Array handling)
- **Cache**: Redis 7
- **DevOps**: Docker & Docker Compose, Makefile
