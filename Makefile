# Bulletin Makefile

# Variables
BINARY_NAME=bulletin-api
GOPATH=$(shell go env GOPATH)
WAILS=$(GOPATH)/bin/wails
WAILS_TAGS=-tags webkit2_41

.PHONY: help setup build build-backend build-frontend build-desktop infra dev dev-backend dev-frontend dev-desktop down db-logs clean

help:
	@echo "Usage:"
	@echo "  make setup            Install dependencies"
	@echo "  make build            Build backend, frontend and desktop"
	@echo "  make build-desktop    Build the Wails desktop client"
	@echo "  make infra            Start infrastructure (Postgres, Redis, Mailhog)"
	@echo "  make dev              Start infrastructure and then run dev commands"
	@echo "  make dev-backend      Start backend with live-reload (Air)"
	@echo "  make dev-frontend     Start frontend with HMR (Vite)"
	@echo "  make dev-desktop      Start Wails in dev mode"
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

build-desktop:
	@echo "--- Building desktop ---"
	$(WAILS) build $(WAILS_TAGS)

infra:
	docker-compose up -d

down:
	docker-compose down

dev-backend:
	cd backend && DATABASE_URL=postgres://bulletin:bulletin_password@localhost:5432/bulletin?sslmode=disable air -c .air.toml

dev-frontend:
	@echo "Starting frontend in dev mode..."
	cd frontend && pnpm run dev

dev-desktop: build-frontend
	@echo "--- Starting desktop in dev mode ---"
	$(WAILS) dev $(WAILS_TAGS)

dev: infra
	@echo "Starting dev environment..."
	@echo "Run 'make dev-backend' and 'make dev-frontend' in separate terminals."
	@echo "Or use a terminal multiplexer like tmux."

db-logs:
	docker-compose logs -f db

install:
	cp backend/bin/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	cp -r frontend/dist/* /var/www/bulletin/

clean:
	rm -rf backend/bin
	rm -rf frontend/dist
	rm -rf frontend/node_modules
