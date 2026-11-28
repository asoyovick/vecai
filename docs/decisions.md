# VECAI Architecture Decisions

## Purpose

This document records important architectural and technical decisions made during the development of VECAI.

The purpose is to preserve the reasoning behind each decision so that future development remains consistent and informed.

Each decision includes:

* Context
* Decision
* Reasoning
* Consequences

---

# Decision 001: Use Go as the Backend Language

## Status

Accepted

## Context

VECAI requires a backend capable of handling concurrent users, API requests, database operations, and future AI integrations.

## Decision

The backend will be developed using Go.

## Reasoning

Go provides:

* Excellent performance
* Built-in concurrency
* Simple syntax
* Strong standard library
* Easy deployment
* Low memory usage
* Excellent support for REST APIs

It also aligns with the developer's learning journey through Zone01.

## Consequences

Advantages

* Fast backend
* Easy deployment
* Scalable architecture
* Strong long-term maintainability

Trade-offs

* Smaller ecosystem than JavaScript
* Steeper learning curve initially

---

# Decision 002: Use the Go Standard Library

## Status

Superseded by Decision 011

## Context

Many Go web frameworks are available.

## Decision

The MVP will use the Go standard library (`net/http`) instead of a third-party web framework.

## Reasoning

Using the standard library encourages a deeper understanding of HTTP, routing, middleware, and request handling.

It also minimizes external dependencies and keeps the project lightweight.

Frameworks can be introduced later if they provide clear benefits.

## Consequences

Advantages

* Better understanding of Go
* Fewer dependencies
* Stable foundation

Trade-offs

* Some features require more manual implementation

## Update

Superseded by [Decision 011](#decision-011-adopt-gin-for-the-http-api) as the API surface grew and route grouping, middleware, and request binding became repetitive with `net/http`.

---

# Decision 003: PostgreSQL as the Primary Database

## Status

Accepted

## Context

VECAI stores structured business data such as users, projects, BOQs, suppliers, quotations, and reports.

## Decision

PostgreSQL will be the primary relational database.

## Reasoning

PostgreSQL provides:

* Reliability
* ACID compliance
* Strong relational modeling
* Excellent performance
* Advanced indexing
* Long-term scalability

## Consequences

Advantages

* Strong data integrity
* Mature ecosystem
* Excellent SQL support

Trade-offs

* Requires schema design before development

---

# Decision 004: Modular Monolith Architecture

## Status

Accepted

## Context

VECAI is initially developed by a solo founder.

## Decision

The application will use a modular monolith architecture.

## Reasoning

A modular monolith is easier to develop, test, deploy, and maintain than microservices.

Business domains remain separated internally while the application is deployed as a single service.

## Consequences

Advantages

* Faster MVP development
* Lower infrastructure costs
* Easier debugging
* Simpler deployment

Trade-offs

* Large applications may eventually require service extraction

---

# Decision 005: HTML Templates for the MVP

## Status

Superseded by Decision 012

## Context

Modern frontend frameworks add significant complexity.

## Decision

The MVP will use server-rendered HTML templates.

## Reasoning

This approach allows rapid feature development while keeping the technology stack simple.

Once the product matures, a separate frontend can be introduced if needed.

## Consequences

Advantages

* Faster development
* Smaller codebase
* Easier deployment

Trade-offs

* Less interactive user interface compared to modern SPAs

## Update

Superseded by [Decision 012](#decision-012-adopt-nextjs-for-the-frontend) as the product moved to a separate frontend application with a richer interactive dashboard and static export/deployment options.

---

# Decision 006: AI Will Not Be Part of the Initial MVP

## Status

Accepted

## Context

The product vision includes AI-assisted construction management.

## Decision

The first release will focus on collecting accurate construction data rather than implementing AI.

## Reasoning

Artificial intelligence depends on reliable data.

The MVP should first provide value through project management, document management, material tracking, and reporting.

Once sufficient data is available, AI features can provide meaningful recommendations.

## Consequences

Advantages

* Faster delivery
* Better data quality
* Practical AI in future releases

Trade-offs

* AI features arrive in later versions

---

# Decision 007: Documentation-Driven Development

## Status

Accepted

## Context

Good software begins with clear understanding before implementation.

## Decision

Major features will be documented before development begins.

Documentation includes:

* Product
* Architecture
* Database
* API
* Roadmap

## Reasoning

Planning reduces rework and improves consistency.

## Consequences

Advantages

* Clear direction
* Easier onboarding
* Better architectural consistency

Trade-offs

* Slightly slower initial planning

---

# Decision 008: Security by Default

## Status

Accepted

## Context

Construction projects often involve sensitive business information.

## Decision

Security considerations will be included from the beginning rather than added later.

## Reasoning

The system will:

* Hash passwords
* Validate user input
* Use HTTPS in production
* Protect sensitive information
* Enforce authorization rules

## Consequences

Advantages

* Better protection of user data
* Reduced security risks

Trade-offs

* Additional implementation effort

---

# Decision 009: Build Incrementally

## Status

Accepted

## Context

The platform has many potential features.

## Decision

Development will focus on one complete feature at a time.

## Reasoning

Delivering small, working increments makes testing, debugging, and user feedback much easier.

## Consequences

Advantages

* Faster feedback
* More stable releases
* Easier debugging

Trade-offs

* Advanced features may take longer to reach users

---

# Future Decisions

As VECAI evolves, this document will be expanded with decisions such as:

* Authentication strategy
* File storage solution
* Cloud deployment platform
* Notification service
* AI model selection
* Mobile application architecture
* Payment integration
* Caching strategy
* Monitoring and logging
* CI/CD pipeline

Each new decision should include its context, reasoning, and expected consequences.

---

# Decision 011: Adopt Gin for the HTTP API

## Status

Accepted

## Context

The backend started with `net/http` (Decision 002), but as the API grew to multiple domain modules, route grouping, middleware composition, JSON binding, and parameter parsing became repetitive.

## Decision

Use [Gin](https://github.com/gin-gonic/gin) as the HTTP framework for the backend API in `backend/cmd/api`.

## Reasoning

Gin provides:

* Route groups for `/api/v1` and domain subroutes
* Middleware for CORS, recovery, logging, and authentication
* Convenient request binding and response helpers
* Good performance and a familiar Go API style

The application still keeps business logic in domain modules under `backend/internal/`, preserving the modular monolith boundary.

## Consequences

Advantages

* Less boilerplate for routing and middleware
* Clearer route registration per domain module
* Faster delivery of new API endpoints

Trade-offs

* Adds a third-party framework dependency
* Some framework conventions must be learned and kept consistent

---

# Decision 012: Adopt Next.js for the Frontend

## Status

Accepted

## Context

The MVP initially planned server-rendered HTML templates (Decision 005). As the product required a richer dashboard, authenticated client flows, and a separate deployable frontend, template rendering in the Go server became limiting.

## Decision

Build the frontend as a separate Next.js application in `frontend/`, using React, TypeScript, and Tailwind CSS.

## Reasoning

Next.js provides:

* A modern component model for dashboard and form-heavy UI
* Client-side routing and authenticated application flows
* A clear separation between the API backend and frontend UI
* Static/export and deployment flexibility

The Go backend remains the API and business-logic layer; the frontend consumes it over HTTP.

## Consequences

Advantages

* Richer, more interactive user experience
* Clear frontend/backend separation
* Frontend can evolve independently

Trade-offs

* Two applications to develop and deploy
* Requires CORS and API contract coordination

---

# Guiding Principle

Every important technical decision should answer one question:

**"Will this make VECAI easier to understand, maintain, and scale over the next five years?"**

If the answer is no, the decision should be reconsidered.

---

# Decision 010: Document File Storage Strategy

## Status

Accepted

## Context

Document uploads were initially stored on the local filesystem (`UPLOAD_DIR`). This is unreliable on ephemeral hosts (e.g. Render) where disk is not persistent across deploys.

## Decision

Introduce a `storage.Storer` interface with a local disk implementation (`LocalStorage`) as the default. The service layer uses the interface for save/open/delete. File uploads enforce an allowlist of MIME types and use server-side content-type sniffing (`http.DetectContentType`) before storage. An S3-compatible implementation can be added later by implementing the same interface (configuration-driven), without changing handlers/services.

## Reasoning

An interface preserves the existing local behavior while enabling durable object storage later. Sniffing + allowlist reduces risk of storing unexpected file types.

## Consequences

Advantages

- Storage backend is swappable (local now, S3-compatible later).
- Basic file-type validation is enforced at upload time.

Trade-offs

- S3 implementation is deferred; production deployments still need durable storage configuration.
- Existing stored files remain on disk until migrated.
