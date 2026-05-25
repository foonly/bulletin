# Bulletin Makefile

.PHONY: help setup infra dev dev-backend dev-frontend build build-backend build-frontend clean db-logs

help:
	@echo "Usage:"
	@echo "  make setup          Install dependencies for backend and frontend"
	@echo "  make infra          Start Postgres and Redis in Docker (detached)"
	@echo "  make dev            Start everything (infra + live-reload backend + vite)"
	@echo "  make build          Build backend and frontend"
	@echo "  make db-logs        Follow database logs"
	@echo "  make clean          Clean build artifacts"

setup:
	@echo "Setting up backend..."
	cd backend && go mod download
	@echo "Installing Air for live-reload..."
	go install github.com/air-verse/air@latest
	@echo "Setting up frontend..."
	cd frontend && pnpm install

infra:
	docker-compose up -d

dev-backend:
	@echo "Starting backend with Air (live-reload)..."
	cd backend && DATABASE_URL=postgres://bulletin:bulletin_password@localhost:5432/bulletin?sslmode=disable air -c .air.toml

dev-frontend:
	@echo "Starting frontend in dev mode..."
	cd frontend && pnpm run dev

dev: infra
	@echo "Starting dev environment..."
	@echo "Run 'make dev-backend' and 'make dev-frontend' in separate terminals."
	@echo "Or use a terminal multiplexer like tmux."

build: build-backend build-frontend

build-backend:
	@echo "Building backend..."
	cd backend && go build -o bin/api ./cmd/api

build-frontend:
	@echo "Building frontend..."
	cd frontend && pnpm run build

db-logs:
	docker-compose logs -f db

clean:
	rm -rf backend/bin
	rm -rf frontend/dist
	rm -rf frontend/node_modules
