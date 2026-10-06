# Architecture

```
Browser
   │  HTTP (same origin in production)
   ▼
Go server (cmd/server)
   ├── /api/*  → internal/routes → internal/handlers
   │                │
   │                ├── middleware (session, role checks, JSON guard)
   │                ├── validation helpers (internal/httpx)
   │                └── database access (SQLite via internal/db)
   └── /*      → static files from dist/ with SPA fallback

Vite dev server (development only)
   ├── serves the React app
   └── proxies /api/* to the Go server on :8080
```

## Layout

```
cmd/server/main.go     entry point: config, db, routes, static hosting
internal/config        environment-based configuration
internal/db            SQLite open + migration runner
internal/auth          scrypt password hashing, session tokens
internal/middleware    session resolution, role checks, JSON CSRF guard
internal/httpx         JSON helpers, validation, row scanning
internal/handlers      one file per API area
internal/routes        endpoint wiring (net/http ServeMux, Go 1.22 patterns)
migrations/            001_init.sql (schema), 002_seed.sql (demo data)
src/                   React + Vite + Tailwind frontend
docs/                  this documentation
```

The Node/TypeScript backend that this application previously used was
fully replaced by the Go backend above; the API contract (routes, JSON
shapes, cookie session) is unchanged so the frontend is untouched in its
data fetching.
