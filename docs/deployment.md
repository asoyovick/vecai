# Deployment

## Backend Docker image

Build and run locally:

```bash
docker build -t vecai-api .
docker run --rm -p 8080:8080 \
  -e DATABASE_URL="postgres://postgres:postgres@host.docker.internal:5433/vecai?sslmode=disable" \
  -e JWT_SECRET="change-me" \
  -e CORS_ALLOWED_ORIGINS="https://your-frontend.example" \
  vecai-api
```

The API runs migrations on startup using `MIGRATIONS_PATH` (default in image: `/app/backend/db/migrations`).

## Render deployment

1. Create a **PostgreSQL** instance on Render and copy its internal/external `DATABASE_URL`.
2. Create a **Web Service** from this repo.
   - Environment: Docker
   - Dockerfile path: `Dockerfile`
3. Set environment variables:
   - `DATABASE_URL` - Render Postgres connection string
   - `JWT_SECRET` - strong random secret
   - `CORS_ALLOWED_ORIGINS` - frontend origin(s), comma-separated
   - `PORT` - provided by Render (the app already reads `PORT`)
   - `UPLOAD_DIR` - persistent disk mount path if using local uploads
4. Migrations: run automatically on startup via `db.RunMigrations`.
5. Uploads: local disk is ephemeral on Render unless a persistent disk is attached. Mount a disk and set `UPLOAD_DIR` to that mount path, or configure object storage (see #15 / storage config) before production use.

## Database backups

Real data requires regular backups.

### Render Postgres

- Enable Render Postgres backups/point-in-time recovery as available on your plan.
- Also keep logical backups:

```bash
pg_dump "$DATABASE_URL" --format=custom --file=vecai-$(date +%Y%m%d-%H%M%S).dump
```

Restore:

```bash
pg_restore --clean --if-exists --no-owner --dbname="$DATABASE_URL" vecai-YYYYMMDD-HHMMSS.dump
```

### Backup routine

- Daily logical `pg_dump` to secure object storage (S3/B2/GCS).
- Retain: 7 daily, 4 weekly, 3 monthly backups.
- Test restore monthly in a staging database.
- Before schema changes or major deploys, take an on-demand backup.
