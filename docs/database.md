# Database

PostgreSQL 16, connected via `DATABASE_URL`
(default `postgres://postgres:vecai@localhost:5432/vecai?sslmode=disable`).

## Migrations

Versioned, reversible migrations live in `db/migrations/` as
`NNNNNN_name.up.sql` / `NNNNNN_name.down.sql` pairs. They are applied in
filename order by `internal/db.Migrate` and tracked in the
`schema_migrations` table. Demo data is loaded from `db/seed.sql`
(`INSERT … ON CONFLICT DO NOTHING`) only when the `users` table is
empty; set `SEED_DEMO=0` to disable.

| # | Migration | Contents |
|---|-----------|----------|
| 000001 | create_projects_table | `projects` (id, name, location, timestamps) |
| 000002 | create_blueprints_and_specs | `blueprints`, `specs` |
| 000003 | create_users_table | `users` (email, password hash, name) |
| 000004 | create_project_members_table | `project_members` |
| 000005 | create_materials_table | `suppliers`, `materials` |
| 000006 | create_daily_reports_table | `daily_reports` |
| 000007 | create_boq_items_table | `boq_items` |
| 000008 | create_project_documents_table | `project_documents` |
| 000009 | create_newsletter_subscribers_table | `newsletter_subscribers` |
| 000010 | add_users_columns | role, job title, initials, color, phone |
| 000011 | add_projects_columns | status enum, progress, dates, client, contractor, budget, description, image |
| 000012 | create_supporting_tables | sessions, organisation, material_deliveries, tasks, daily_report_materials, budget_months, activity_log, user_preferences |

## Tables

- `users` — accounts with role (`admin` | `member` | `viewer`), job title,
  initials, color, phone. Passwords stored as `scrypt:salt:hash`.
- `sessions` — server-side session tokens (SHA-256 of the cookie value),
  7-day expiry.
- `organisation` — single-row workspace settings.
- `projects` — status in a fixed enum, progress, budget, dates, client,
  contractor, description, image.
- `project_members` — many-to-many projects ↔ users.
- `suppliers` — vendor directory with rating and status.
- `materials` — per-project stock: required/available/used quantities,
  status (`ok|low|critical|ordered`), unit cost.
- `material_deliveries` — each delivery raises on-hand quantity and
  recomputes status.
- `tasks` — project tasks with assignee, dates, status, priority.
- `project_documents` — uploaded files with category enum, version,
  stored name. Seeded documents are records only (`stored_name` begins
  with `seeded/`); uploaded files land in `UPLOADS_DIR`
  (default `data/uploads`).
- `daily_reports` + `daily_report_materials` — daily site diaries
  (served at `/api/site-updates`).
- `boq_items` — bill-of-quantities line items; `amount` is computed as
  `quantity * unit_rate` by the API.
- `blueprints` / `specs` — blueprint and specification records.
- `newsletter_subscribers` — public newsletter signups.
- `budget_months` — monthly planned vs. actual per project.
- `activity_log` — audit trail of every mutation.
- `user_preferences` — per-user notification preferences (JSON).

## Conventions

- Timestamps are `TEXT` columns in `YYYY-MM-DD HH24:MI:SS` UTC format
  (the format the frontend has always rendered), with
  `to_char(now() AT TIME ZONE 'UTC', …)` defaults.
- All foreign keys are `ON DELETE CASCADE` where child rows are owned by
  a parent, and `ON DELETE SET NULL` where the actor is merely
  referenced.
- Serial sequences are reset by the seed so demo ids stay stable.

## Running against a fresh database

```bash
docker run --rm -p 5432:5432 \
  -e POSTGRES_PASSWORD=vecai -e POSTGRES_DB=vecai postgres:16-alpine
DATABASE_URL="postgres://postgres:vecai@localhost:5432/vecai?sslmode=disable" \
  go run ./cmd/api
```
