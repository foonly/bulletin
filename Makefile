# Bulletin Makefile

.PHONY: help setup build up down infra dev dev-backend dev-frontend clean db-logs

help:
	@echo "Usage:"
	@echo "  make setup          Install dependencies for backend and frontend"
	@echo "  make infra          Start Postgres and Redis in Docker (detached)"
	@echo "  make dev            Start everything (infra + live-reload backend + vite)"
	@echo "  make down           Stop infrastructure"
	@echo "  make db-logs        Follow database logs"
	@echo "  make clean          Clean build artifacts"

setup:
	@echo "Setting up backend..."
	cd backend && go mod download
	@echo "Installing Air for live-reload..."
	go install github.com/air-verse/air@latest
	@echo "Setting up frontend..."
	cd frontend && npm install

build: build-backend build-frontend

build-backend:
	@echo "Building backend binary..."
	cd backend && go build -o bin/bulletin-api cmd/api/*.go

build-frontend:
	@echo "Building frontend assets..."
	cd frontend && npm run build

infra:
	docker-compose up -d

dev-backend:
	@echo "Starting backend with Air (live-reload)..."
	cd backend && DATABASE_URL=postgres://bulletin:bulletin_password@localhost:5432/bulletin?sslmode=disable air -c .air.toml

dev-frontend:
	@echo "Starting frontend in dev mode..."
	cd frontend && npm run dev

dev: infra
	@echo "Starting dev environment..."
	@echo "Run 'make dev-backend' and 'make dev-frontend' in separate terminals."
	@echo "Or use a terminal multiplexer like tmux."

db-logs:
	docker-compose logs -f db

clean:
	rm -rf backend/bin
	rm -rf frontend/dist
	rm -rf frontend/node_modules
