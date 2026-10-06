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
- **BOQ** — bill-of-quantities line items with computed amounts.
- **Reports** — weekly status, budget variance and material alerts.
- **Project intelligence** — a rule-based assistant that answers questions
  from live project data (weekly reports, site summaries, budget variance,
  material forecast, risk analysis).

## Architecture

- `src/` — React 19 + Vite + Tailwind CSS v4 frontend (public landing page,
  login/register, authenticated workspace under `/app`)
- `cmd/api` — Go entry point: configuration and dependency wiring only
- `internal/<domain>/` — domain-oriented backend (auth, project, materials,
  boq, documents, reports, newsletter); each domain owns its handler,
  service, repository interface and PostgreSQL implementation
- `internal/api` — shared HTTP concerns (CORS, JSON responses, guards)
- `internal/db` — PostgreSQL connection pool and migration runner
- `db/migrations` — versioned, reversible SQL migrations
- `db/seed.sql` — demo data, loaded automatically on an empty database
- `data/` — runtime uploads directory (git-ignored)
- `docs/` — product and technical documentation

## Development

PostgreSQL 16 and the Vite dev server run side by side; Vite proxies
`/api` to the Go server.

```bash
# Terminal 1 — PostgreSQL (docker)
docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=vecai -e POSTGRES_DB=vecai postgres:16-alpine

# Terminal 2 — Go API
DATABASE_URL="postgres://postgres:vecai@localhost:5432/vecai?sslmode=disable" go run ./cmd/api

# Terminal 3 — frontend
pnpm dev
```

Open http://localhost:8443. Demo sign-in: `david.ochieng@vecai.co.ke` /
`demo1234` (all seeded accounts share this password).

## Production build

```bash
pnpm build
DATABASE_URL="postgres://…" STATIC_DIR=dist go run ./cmd/api
```

The Go server then serves both the built frontend (SPA fallback) and the
API on `:8080`.

## Testing

```bash
gofmt -w ./cmd ./internal
go vet ./...
go test ./...                      # unit tests
go test -count=1 ./cmd/...         # API integration tests (need a test database)
```

Integration tests use `postgres://postgres:vecai@localhost:5432/vecai_test`
and recreate its schema on every run.

## Documentation

`docs/` contains the product overview, architecture, database schema,
API reference and roadmap. See `docs/architecture.md` for the request
flow through the domain packages.
