# VECAI

VECAI (Victoria Eagles Construct AI) is a construction project management
platform for Kenyan construction teams. It replaces spreadsheets, WhatsApp
groups and paper site diaries with one workspace for projects, tasks,
materials, documents, daily site reports, budgets and suppliers.

## What it does

- **Projects** — track status, progress, budget and team for every project.
- **Tasks** — assign, schedule and track tasks by project.
- **Materials** — stock levels, supplier links and delivery logging.
- **Documents** — versioned project documents of every category.
- **Site updates** — structured daily site diaries from the field.
- **Reports** — weekly status, budget variance and material alerts.
- **Project intelligence** — a rule-based assistant that answers questions
  from live project data (weekly reports, site summaries, budget variance,
  material forecast, risk analysis).

## Architecture

- `src/` — React 19 + Vite + Tailwind CSS v4 frontend
- `cmd/server` — Go entry point
- `internal/` — config, database, auth, handlers, middleware, routes
- `migrations/` — SQL migrations (schema + demo seed)
- `data/` — runtime SQLite database and uploaded files (git-ignored)
- `docs/` — product and technical documentation

## Development

The Go server (API on `:8080`) and the Vite dev server (UI on `$PORT`,
default 8443) run side by side; Vite proxies `/api` to the Go server.

```bash
# Terminal 1 — Go API
DATABASE_PATH=data/vecai.db go run ./cmd/server

# Terminal 2 — frontend
pnpm dev
```

## Production build

```bash
pnpm build
DATABASE_PATH=data/vecai.db STATIC_DIR=dist go run ./cmd/server
```

The Go server then serves both the built frontend and the API on `:8080`.

## Demo credentials

`david.ochieng@vecai.co.ke` / `demo1234` (admin). Seeded accounts share the
same password; see the login page.

## Testing

```bash
go test ./...
npx tsc --noEmit
pnpm build
```
