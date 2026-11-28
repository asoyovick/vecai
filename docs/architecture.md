# VECAI System Architecture

## 1. Overview

VECAI (Victoria Eagles Construct AI) is a construction project management platform with a Go API backend and a separate Next.js frontend.

The MVP follows a **modular monolithic architecture** on the backend: one deployable Go service contains domain modules with clear boundaries. The frontend is a separate Next.js application that communicates with the backend over HTTP/JSON.

---

## 2. Architecture Principles

* Simplicity before complexity
* Modular design and separation of concerns
* Idiomatic Go in the backend
* Secure by default
* Database-driven development
* API-first backend design
* Scalability for future AI integration

---

## 3. High-Level Architecture

```text
Browser / User
      │
      ▼
Next.js Frontend (frontend/)
      │  HTTP/JSON
      ▼
Gin API Server (backend/cmd/api)
      │
      ├── Middleware: Recovery, Logger, CORS, Optional/Auth JWT
      │
      ├── Domain Handlers / Services / Repositories
      │   ├── Auth
      │   ├── Projects / Members / Blueprints / Specs
      │   ├── Materials
      │   ├── Daily Reports
      │   ├── Documents
      │   ├── BOQ
      │   └── Newsletter
      │
      ├── PostgreSQL (persistent data)
      └── File Storage (local disk via storage interface; S3-compatible planned)
```

---

## 4. System Components

### 4.1 Frontend

Location: `frontend/`

* Next.js (App Router)
* React + TypeScript
* Tailwind CSS
* Framer Motion / UI components for dashboard and marketing pages

The frontend calls the backend API and manages authenticated user flows in the browser.

### 4.2 API Server

Location: `backend/cmd/api/main.go`

The Gin HTTP server is responsible for:

* Receiving HTTP requests
* CORS handling
* Request logging and panic recovery
* Authentication middleware (optional/authenticated JWT middleware)
* Route registration
* JSON responses

Handlers parse and validate input, then delegate to services.

### 4.3 Business Logic Layer

Location: `backend/internal/<domain>/`

Each domain module typically contains:

* `handler.go` / route registration - HTTP transport
* `service.go` - business rules and orchestration
* `repository.go` / `postgres_repo.go` - data access
* `model.go` - domain types

Business rules are kept out of `main.go` and out of the frontend.

### 4.4 Data Layer

Primary database: PostgreSQL.

Migrations: `backend/db/migrations/` (golang-migrate format, 11 migrations currently).

Database access is performed through repository interfaces/implementations in domain modules. The backend runs migrations on startup.

### 4.5 File Storage

Document files are stored through a `storage.Storer` interface (`backend/internal/storage/`). The default implementation stores files on local disk under `UPLOAD_DIR`. Document metadata is stored in PostgreSQL. An S3-compatible implementation is planned.

---

## 5. Current Domain Modules

| Module | Responsibility | Main routes |
| --- | --- | --- |
| Auth | Register/login/verify/logout, current user profile | `/api/v1/auth/*`, `/api/v1/users/me` |
| Project | Projects, members, blueprints, specs | `/api/v1/projects/*` |
| Materials | Material delivery records and summary | `/api/v1/projects/:id/materials*` |
| Reports | Daily site reports and summary | `/api/v1/projects/:id/reports*` |
| Documents | Project document metadata and file upload/download | `/api/v1/projects/:id/documents*` |
| BOQ | Bill of quantities items and comparison | `/api/v1/projects/:id/boq*` |
| Newsletter | Newsletter subscriptions | `/api/v1/newsletter/subscribe` |

Planned modules not yet implemented as full domain slices include suppliers, quotations, notifications, timeline/activity, and search.

---

## 6. Request Flow

```text
User action
  ↓
Next.js frontend
  ↓
HTTP request to /api/v1/...
  ↓
Gin middleware (CORS, recovery, logger, auth)
  ↓
Domain handler (bind/validate request)
  ↓
Domain service (business rules)
  ↓
Repository / storage
  ↓
PostgreSQL or file storage
  ↓
JSON response envelope
  ↓
Frontend UI update
```

---

## 7. Security Architecture

Current controls:

* Password hashing in auth service
* JWT-based authentication middleware
* Role-aware authorization checks in domain handlers/services where implemented
* Server-side request validation/binding
* Parameterized SQL queries through database/sql
* CORS allowlist configured by `CORS_ALLOWED_ORIGINS`
* Upload content-type sniffing and allowlist for document uploads

Planned/hardening:

* HTTPS in production deployment
* CSRF strategy appropriate for token-based frontend auth
* Rate limiting
* Audit logging and monitoring
* Production-grade secret management

---

## 8. AI Integration Strategy

Artificial Intelligence is **not part of the current MVP implementation**. The MVP focuses on collecting reliable project data.

Future AI services may analyze stored data for:

* Material demand prediction
* Cost forecasting
* Productivity analysis
* Schedule recommendations
* Risk detection

AI should act as decision support, not a replacement for construction professionals.

---

## 9. Scalability Strategy

The backend remains a modular monolith while product scope is evolving. As adoption grows, modules can be extracted into services if needed.

Potential future extractions/services:

* AI/analytics service
* Notification service
* Reporting/exports service
* Object storage/CDN-backed document service
* Payment/billing service

---

## 10. Technology Stack

### Backend

* Go
* Gin (`github.com/gin-gonic/gin`)
* PostgreSQL
* `database/sql` + `lib/pq`
* `golang-migrate` for migrations
* JWT (`github.com/golang-jwt/jwt/v5`)

### Frontend

* Next.js
* React
* TypeScript
* Tailwind CSS

### Development

* Git / GitHub
* Makefile for backend/dev tasks
* Air for optional live reload
* Docker planned for deployment workflows

---

## 11. Architectural Goals

The architecture aims for maintainability, simplicity, performance, security, reliability, and extensibility. The primary objective is a production-ready MVP that can evolve without a complete redesign.
