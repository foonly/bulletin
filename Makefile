# Bulletin Makefile

# Variables
BINARY_NAME=bulletin-api

.PHONY: help setup build build-backend build-frontend infra dev dev-backend dev-frontend down db-logs clean

help:
	@echo "Usage:"
	@echo "  make setup            Install dependencies"
	@echo "  make build            Build backend and frontend"
	@echo "  make infra            Start infrastructure (Postgres, Redis, Mailhog)"
	@echo "  make dev              Start infrastructure and then run dev commands"
	@echo "  make dev-backend      Start backend with live-reload (Air)"
	@echo "  make dev-frontend     Start frontend with HMR (Vite)"
	@echo "  make down             Stop infrastructure"
	@echo "  make db-logs          Follow database logs"
	@echo "  make clean            Remove build artifacts and dependencies"

setup:
	@echo "--- Setting up project ---"
	cd backend && go mod download
	go install github.com/air-verse/air@latest
	@echo "Setting up frontend..."
	cd frontend && pnpm install

build: build-backend build-frontend

build-backend:
	@echo "--- Building backend ---"
	mkdir -p backend/bin
	cd backend && go build -o bin/$(BINARY_NAME) cmd/api/*.go

build-frontend:
	@echo "--- Building frontend ---"
	cd frontend && pnpm run build

infra:
	docker-compose up -d

down:
	docker-compose down

dev-backend:
	cd backend && DATABASE_URL=postgres://bulletin:bulletin_password@localhost:5432/bulletin?sslmode=disable air -c .air.toml

dev-frontend:
	@echo "Starting frontend in dev mode..."
	cd frontend && pnpm run dev

dev: infra
	@echo "Starting dev environment..."
	@echo "Run 'make dev-backend' and 'make dev-frontend' in separate terminals."
	@echo "Or use a terminal multiplexer like tmux."

build: build-backend build-frontend

db-logs:
	docker-compose logs -f db

clean:
	rm -rf backend/bin
	rm -rf frontend/dist
	rm -rf frontend/node_modules
