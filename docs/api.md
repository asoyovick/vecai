# API

Base URL: same origin (`/api`). All responses are JSON. Errors are
`{ "error": "message" }` with an appropriate status code. Auth is a single
`HttpOnly` session cookie (`vecai_session`).

## Auth

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/auth/login | `{email, password}` → `{user}` + cookie |
| POST | /api/auth/logout | destroys the session |
| GET  | /api/auth/me | current user |

## Projects

| Method | Path | Description |
|--------|------|-------------|
| GET   | /api/projects | list (any signed-in user) |
| POST  | /api/projects | create (admin) |
| GET   | /api/projects/{id} | project detail incl. tasks, materials, documents, activity, budget |
| PATCH | /api/projects/{id} | update name/location/status/progress/budget/description (admin) |
| GET   | /api/projects/{id}/budget | monthly budget rows |

## Tasks

| Method | Path | Description |
|--------|------|-------------|
| GET   | /api/tasks?projectId&status&q | list |
| POST  | /api/tasks | create (not viewer) |
| PATCH | /api/tasks/{id} | update status/priority/name/description/assignee/due date (not viewer) |

## Materials

| Method | Path | Description |
|--------|------|-------------|
| GET  | /api/materials?projectId&status&q | list |
| POST | /api/materials | add (not viewer) |
| POST | /api/materials/{id}/deliveries | log a delivery, updates on-hand + status (not viewer) |

## Documents

| Method | Path | Description |
|--------|------|-------------|
| GET  | /api/documents?projectId&category&q | list |
| POST | /api/documents | multipart upload `{file, projectId, name, category, version}` (not viewer) |
| GET  | /api/documents/{id}/download | file download |

## Site updates

| Method | Path | Description |
|--------|------|-------------|
| GET  | /api/site-updates?projectId | list with materials used |
| POST | /api/site-updates | submit daily report (not viewer) |

## Workspace

| Method | Path | Description |
|--------|------|-------------|
| GET  | /api/dashboard | stats, projects, alerts, deadlines, activity |
| GET  | /api/activity?projectId | activity feed |
| GET  | /api/team | members + assignments |
| POST | /api/team/invite | invite member (admin) |
| GET  | /api/suppliers?category&q | list |
| POST | /api/suppliers | add (not viewer) |
| PATCH| /api/me | update own profile |
| POST | /api/me/password | change password; invalidates other sessions |
| GET/PATCH | /api/me/preferences | notification preferences |
| GET/PATCH | /api/organisation | workspace settings (patch: admin) |

## Project intelligence

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/assistant | `{question, projectId?}` → computed answer |
