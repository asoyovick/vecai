# Sprint 01

Scope of the first production sprint:

1. Public landing page at `/` before authentication.
2. URL routing: `/` landing, `/login`, authenticated app under `/app`.
3. Real authentication backed by the Go API (scrypt passwords,
   server-side sessions, HttpOnly cookie).
4. Complete Go backend replacing the embedded Node one — same
   endpoints, same JSON shapes.
5. SQLite migrations + seeded demo data.
6. Frontend fully wired to the Go API (no mock data, no fake APIs).
7. Tests: Go unit/integration tests; `tsc` clean; Vite build green.

Out of scope (see roadmap): email sending, password reset, real photo
uploads, PostgreSQL.
