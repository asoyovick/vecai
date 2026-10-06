# Architecture

```
Browser
   │  HTTP (same origin in production; Vite proxies /api in development)
   ▼
Go server (cmd/api)
   ├── /api/auth/*        → internal/auth
   ├── /api/projects/*    → internal/project   (+ /api/tasks)
   ├── /api/materials/*   → internal/materials (+ /api/suppliers)
   ├── /api/documents/*   → internal/documents
   ├── /api/boq/*         → internal/boq
   ├── /api/dashboard, /api/activity, /api/site-updates,
   │  /api/assistant      → internal/reports
   ├── /api/newsletter/*  → internal/newsletter
   └── /*                 → static files from dist/ with SPA fallback

PostgreSQL
   ▲
   │
internal/db  (pool + migration runner over db/migrations)
```

## Domain-oriented layout

```
cmd/api/main.go            configuration + dependency wiring
db/migrations/             000001…000012 up/down migrations
db/seed.sql                demo data (loaded on an empty database)
internal/api/cors.go       origin allowlist (VECAI_CORS_ORIGINS)
internal/api/response.go   JSON responses, decode guard, panic recovery,
                           shared field validation
internal/auth/             registration, login, sessions, password policy,
                           user management, team invites, middleware
internal/project/          projects, budget months, tasks
internal/materials/        materials, deliveries, suppliers
internal/documents/        upload, listing, download
internal/boq/              bill-of-quantities items
internal/reports/          dashboard, activity feed, daily reports,
                           rule-based assistant
internal/newsletter/       public newsletter subscriptions
internal/db/               PostgreSQL pool, migration runner, row helpers
```

Every domain follows the same layering:

```
handler.go      HTTP concerns: parse requests, validate input, call the
                service, map errors to status codes
service.go      business rules: validation, coordination, transactions
repository.go   the interface the service depends on
postgres_repo.go  the PostgreSQL implementation of that interface
model.go        domain constants and types
```

The dependency direction is one-way:

```
handler → service → repository interface → postgres repository → PostgreSQL
```

Handlers never contain business logic and never execute SQL. Services
never import `net/http`. Domains are self-contained; shared concerns
(CORS, JSON formatting, validation helpers) live in `internal/api`.

## Request flow

```
Frontend
   ↓  HTTP API (JSON, vecai_session cookie)
Go handler        — decode + validate the request
Go service        — enforce business rules, authorise the action
Repository        — interface the service depends on
PostgreSQL        — via the domain's postgres_repo implementation
   ↓  same path back up
JSON response     — camelCase for resource payloads, snake_case where
                    the frontend has always consumed raw rows
Frontend
```

## Authentication and authorization

- Registration (`POST /api/auth/register`) creates a `member` account.
- Login verifies a scrypt hash (`scrypt:salt:hash`, salt is the hex
  string) and issues a server-side session token in an HttpOnly,
  SameSite=Lax cookie (`vecai_session`), 7-day expiry.
- `internal/auth.Middleware` resolves the cookie to a user on the
  request context; `RequireUser` / `RequireRole` enforce it.
- Roles: `admin` (creates projects, invites members, edits
  organisation), `member` (working user), `viewer` (read-only —
  blocked from every mutation endpoint).
- Services re-check the role on every mutation; the frontend never
  supplies authorization data the server trusts.

## Frontend

`src/App.tsx` owns a small URL router (`routeFromPath` / `pathForRoute`
with `pushState` + `popstate`):

- `/` — public landing page (with newsletter signup)
- `/login`, `/register` — public auth pages
- `/app/*` — authenticated workspace; project deep links
  (`/app/projects/:id`) are supported

`vite.config.ts` proxies `/api` to the Go server (`VECAI_API_URL`
overrides the default `http://localhost:8080`), so development and
production share one API contract.
