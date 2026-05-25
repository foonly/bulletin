# Development Guide

## Prerequisites

- **Docker and Docker Compose**: For running infrastructure.
- **Go 1.22+**: For running the backend natively.
- **Node.js 20+**: For running the frontend natively.

## Nimble Setup (Recommended)

To achieve the fastest feedback loop, we run infrastructure in Docker and application code on the host machine.

1. **Install Dependencies**:

   ```bash
   make setup
   ```

   _Note: This also installs `air` for Go live-reloading._

2. **Start Infrastructure**:

   ```bash
   make infra
   ```

   _Starts Postgres, Redis, and Mailhog in the background._

3. **Email Testing**:
   Access the Mailhog Web UI at `http://localhost:8025` to view all outgoing system emails (resets, verifications). No SMTP authentication is required for Mailhog in the default setup.

4. **Run Application**:
   Open two terminals:
   - Terminal 1: `make dev-backend` (Live-reloads on `.go` or `.sql` change)
   - Terminal 2: `make dev-frontend` (Vite HMR)

## Environment Variables

### Frontend (.env)

- `VITE_SITE_NAME`: The name displayed in headers and browser tabs.

### Backend (System Environment)

- `PORT`: The port the API will listen on (default: 8080).
- `DATABASE_URL`: Postgres connection string.
- `SMTP_HOST`: SMTP server hostname (default: localhost).
- `SMTP_PORT`: SMTP server port (default: 1025).
- `SMTP_FROM`: The email address appearing in the "From" field.
- `SMTP_USER`: SMTP authentication username (optional).
- `SMTP_PASS`: SMTP authentication password (optional).
- `SMTP_USE_TLS`: Set to `true` for implicit TLS (usually port 465). Standard STARTTLS (port 587) is supported automatically.
- `FRONTEND_URL`: The base URL of the frontend (e.g., `https://bulletin.example.com`).

## First-time Registration

The system bootstraps an initial circle and an invite code `welcome`.

1. Open `http://localhost:5173/register`
2. Use invite code `welcome`.
3. This user becomes the first `admin`.

## Database Management

- **Migrations**: Add new `.up.sql` files to the `migrations` folder. They are tracked in `schema_migrations` and applied automatically when the backend starts.
- **Reset**: To wipe everything and start fresh:
  ```bash
  docker-compose down -v
  ```

## Production Deployment (Ubuntu 24.04)

### 1. Install Dependencies

```bash
sudo apt update
sudo apt install postgresql postgresql-contrib redis-server nginx
```

### 2. Configure PostgreSQL

Create the database and user matching the default `DATABASE_URL`:

```bash
sudo -u postgres psql -c "CREATE USER bulletin WITH PASSWORD 'bulletin_password';"
sudo -u postgres psql -c "CREATE DATABASE bulletin OWNER bulletin;"
```

### 3. Install API Service

1. Build the binary: `make build-backend`
2. Copy binary: `sudo cp backend/bin/bulletin-api /usr/local/bin/`
3. Install service: `sudo cp bulletin-api.service /etc/systemd/system/`
4. Start: `sudo systemctl enable --now bulletin-api`

### 4. Configure Web Server

1. Build frontend: `make build-frontend`
2. Copy assets: `sudo cp -r frontend/dist/* /var/www/bulletin/frontend/`
3. Install Nginx config: `sudo cp nginx.conf /etc/nginx/sites-available/bulletin`
4. Enable site: `sudo ln -s /etc/nginx/sites-available/bulletin /etc/nginx/sites-enabled/`
5. Restart Nginx: `sudo systemctl restart nginx`

## UI & Notifications

Bulletin uses a custom **Toast System**. Avoid using standard `alert()` calls.
Instead, use the `toast` store in your components:

```javascript
const toast = useToastStore();
toast.success("Done!");
toast.error("Failed", 20000); // 20s duration
```

## Route Security

The backend uses a `MembershipMiddleware`. If you add a new circle-scoped endpoint in `main.go`, ensure it is registered within a group that uses this middleware to prevent unauthorized data access.
