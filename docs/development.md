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

   _Starts Postgres and Redis in the background._

3. **Run Application**:
   Open two terminals:
   - Terminal 1: `make dev-backend` (Live-reloads on `.go` or `.sql` change)
   - Terminal 2: `make dev-frontend` (Vite HMR)

## First-time Registration

The system bootstraps an initial circle and an invite code `welcome`.

1. Open `http://localhost:5173/register`
2. Use invite code `welcome`.
3. This user becomes the first `admin`.

## Database Management

- **Migrations**: Add new `.up.sql` files to the `migrations` folder. They are applied automatically when the backend starts.
- **Reset**: To wipe everything and start fresh:
  ```bash
  docker-compose down -v
  ```

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
