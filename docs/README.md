# Bulletin Documentation

Welcome to the Bulletin technical documentation. Bulletin is a communications system based on social circles, featuring threaded posts and real-time chat.

## Table of Contents

1. [Architecture Overview](./architecture.md)
2. [Database Schema](./database.md)
3. [API Reference](./api.md)
4. [Development Guide](./development.md)
5. [Invite System & Permissions](./invites.md)

## Project Goals

- **Privacy**: Circles are only visible to their members.
- **Organization**: Communication is split between persistent threaded posts (organized by tags) and real-time chat.
- **Traceability**: All users and circle memberships are tracked via an invite chain.
- **Retention**: Configurable chat history limits (default 50 messages or 14 days).

## Technology Stack

- **Backend**: Go (chi, pgx, gorilla/websocket)
- **Frontend**: Vue 3 (Pinia, Vue Router, Tailwind CSS v4)
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Deployment**: Docker & Docker Compose
