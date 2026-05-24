# Development Guide

## Prerequisites

- Docker and Docker Compose
- Go 1.22+ (for local development)
- Node.js 20+ (for local development)

## Fast Start

Run the entire stack in Docker:
```bash
make up
```

## Initial Setup

1. **Install Dependencies**:
   ```bash
   make setup
   ```

2. **First-time Registration**:
   The system bootstraps an initial circle and an invite code `welcome`.
   1. Open `http://localhost:5173/register`
   2. Use invite code `welcome`.
   3. This first user is granted `admin` rights in the default circle.

## Backend Development

The backend uses a standard Go layout.
- `cmd/api`: Main entry point and server setup.
- `internal/auth`: Session middleware and authentication handlers.
- `internal/chat`: WebSocket Hub and real-time logic.
- `internal/posts`: Core business logic for circles and posts.

To run locally with a Docker-hosted database:
```bash
docker-compose up -d db redis
make dev-backend
```

## Frontend Development

The frontend is a Vue 3 SPA using Vite.
- `src/stores`: Pinia state management.
- `src/views`: Top-level page components.
- `src/router`: Navigation logic.

Run the dev server:
```bash
make dev-frontend
```
Vite is configured to proxy `/api` requests to `localhost:8080`.
