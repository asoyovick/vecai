# VECAI

**VECAI (Victoria Eagles Construct AI)** is a construction project management platform designed to help contractors, engineers, architects, quantity surveyors, suppliers, and project owners manage construction work more efficiently.

The platform centralizes project information, tracks materials, manages documents, and supports daily site reporting. Future versions will add AI-assisted forecasting and decision support.

---

## Vision

To become a leading intelligent construction management platform that enables construction professionals to deliver projects efficiently, transparently, and with minimal resource wastage.

## Mission

To digitize construction project management by providing a practical platform that improves collaboration, accountability, productivity, and decision-making.

---

## Current Status

🚧 **Under active development (MVP).**

Implemented or in-progress foundations include:

* Authentication and current-user profile endpoints
* Project CRUD, project members, blueprints, and specs
* Material tracking and summaries
* Daily reports and summaries
* BOQ items and comparison
* Project document upload/download metadata flow
* Newsletter subscription endpoint
* Next.js frontend with authentication and project dashboard flows

Planned next: daily reports UI, CI, suppliers/quotations, notifications, timeline, search, and deployment hardening. See [docs/roadmap.md](docs/roadmap.md) and open GitHub issues.

---

## Technology Stack

### Backend

* Go
* Gin (`github.com/gin-gonic/gin`)
* PostgreSQL
* `database/sql` + `lib/pq`
* `golang-migrate/migrate`
* JWT authentication (`github.com/golang-jwt/jwt/v5`)

### Frontend

* Next.js (App Router)
* React
* TypeScript
* Tailwind CSS

### Development Tools

* Git / GitHub
* Make
* Air (optional live reload)
* Docker (planned)

---

## Project Structure

```text
vecai/
├── backend/
│   ├── cmd/api/              # API entrypoint
│   ├── db/migrations/        # SQL migrations (golang-migrate)
│   └── internal/             # Domain modules and shared backend packages
│       ├── api/              # Shared HTTP helpers (CORS, response envelope)
│       ├── auth/             # Authentication and users
│       ├── boq/              # Bill of quantities
│       ├── db/               # DB connection and migration runner
│       ├── documents/        # Project documents
│       ├── materials/        # Material tracking
│       ├── newsletter/       # Newsletter subscriptions
│       ├── project/          # Projects, members, blueprints, specs
│       ├── reports/          # Daily reports
│       └── storage/          # File storage abstraction
├── docs/                     # Product and engineering documentation
├── frontend/                 # Next.js frontend application
├── .air.toml                 # Air live-reload config
├── go.mod
├── Makefile
└── README.md
```

---

## Getting Started

### Prerequisites

* Go (version from `go.mod`)
* Node.js + npm
* PostgreSQL
* Make

### Backend

Default backend config (can be overridden by env vars):

* `PORT=8080`
* `DATABASE_URL=postgres://postgres:postgres@localhost:5433/vecai?sslmode=disable`
* `MIGRATIONS_PATH=backend/db/migrations`
* `UPLOAD_DIR=backend/uploads`
* `JWT_SECRET=vecai-default-dev-secret-change-in-prod`

Run with Make:

```bash
make run
```

For live reload (if `air` is installed):

```bash
make dev
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend runs on Next.js dev server (default `http://localhost:3000`) and calls the API.

---

## Documentation

Project documentation lives in [`docs/`](docs/):

* [product.md](docs/product.md) - Product vision and scope
* [architecture.md](docs/architecture.md) - System architecture
* [database.md](docs/database.md) - Database schema and migration overview
* [api.md](docs/api.md) - HTTP API endpoints
* [roadmap.md](docs/roadmap.md) - Development roadmap
* [sprint-01.md](docs/sprint-01.md) - Sprint 01 notes
* [decisions.md](docs/decisions.md) - Architecture decision records
* [coding-standards.md](docs/coding-standards.md) - Coding conventions
* [glossary.md](docs/glossary.md) - Domain glossary

---

## Development Principles

* Simplicity first
* Modular monolith backend with clear domain boundaries
* Secure by default
* Testable, maintainable code
* Incremental MVP delivery

---

## Author

**Victor Ouma**  
Founder, Victoria Eagles Brand

## License

This project is currently proprietary. All rights reserved by Victoria Eagles Brand.
