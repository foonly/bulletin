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
- **Markdown Support**: Full GFM support for posts and chat with real-time previews and strict HTML sanitization (DOMPurify).
- **Multi-Factor Auth**: TOTP-based 2FA support for enhanced account security.
- **Email Integration**: Integrated mailer for password resets and email verification (tested via Mailhog).
- **Advanced Dashboard**: Visual overview of all circles with real-time unread counts and activity tracking.
- **Search**: Cross-thread search within circles to quickly find historical content.
- **Refined Routing**: Fully navigable deep-links for every part of the circle (Posts, Chat, Search, Settings).
- **Schema Migrations**: State-aware migration runner for reliable infrastructure updates.
- **Post Deletion**: Secure deletion system that preserves thread integrity for replies.
- **Unified Sidebar**: Consolidated navigation with collapsible member lists and tag-based activity badges.

## Technology Stack

- **Backend**: Go 1.25 (chi, pgx, gorilla/websocket, totp, air)
- **Frontend**: Vue 3 (Pinia, Vue Router, Tailwind CSS v4, marked, dompurify)
- **Email Testing**: Mailhog (SMTP)
- **Database**: PostgreSQL 16 (Recursive CTEs, Migration tracking)
- **Cache**: Redis 7
- **DevOps**: Docker & Docker Compose, Makefile
