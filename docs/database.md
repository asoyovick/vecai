# Database

SQLite (WAL mode) at `data/vecai.db`, overridable with `DATABASE_PATH`.

## Tables

- `users` — accounts with role (`admin` | `member` | `viewer`), job title,
  initials, color, phone. Passwords stored as `scrypt:salt:hash`.
- `sessions` — server-side session tokens (SHA-256 of the cookie value),
  7-day expiry.
- `organisation` — single-row workspace settings.
- `projects` — status in a fixed enum, progress, budget, dates, client,
  contractor, description.
- `project_members` — many-to-many projects ↔ users.
- `suppliers` — vendor directory with rating and status.
- `materials` — per-project stock: required/available/used quantities,
  status (`ok|low|critical|ordered`), unit cost.
- `material_deliveries` — each delivery raises on-hand quantity and
  recomputes status.
- `tasks` — project tasks with assignee, dates, status, priority.
- `documents` — uploaded files with category enum, version, stored name.
- `site_updates` + `site_update_materials` — daily site diaries.
- `budget_months` — monthly planned vs. actual per project.
- `activity_log` — audit trail of every mutation.
- `user_preferences` — per-user notification preferences (JSON).

Migrations live in `migrations/*.sql` and are applied in filename order,
tracked in the `schema_migrations` table. `002_seed.sql` seeds demo
projects, users and data using `INSERT OR IGNORE`, so it is safe on an
existing database.

All foreign keys are `ON DELETE CASCADE` where child rows are owned by a
parent, and `ON DELETE SET NULL` where the actor is merely referenced.
