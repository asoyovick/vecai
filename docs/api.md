# API

Base URL: same origin (`/api`). All responses are JSON. Errors are
`{ "error": "message" }` with an appropriate status code. Auth is a single
`HttpOnly` session cookie (`vecai_session`).

## Auth

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/auth/register | `{name, email, password, jobTitle?}` → `{user}` + cookie (public) |
| POST | /api/auth/login | `{email, password}` → `{user}` + cookie |
| POST | /api/auth/logout | destroys the session |
| GET  | /api/auth/me | current user |

Registration and login validate at the API boundary: unique email,
valid address, password of 8–200 characters with at least one letter
and one number. Duplicate emails return `409`.

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

## BOQ

| Method | Path | Description |
|--------|------|-------------|
| GET   | /api/boq?projectId | list items with computed `amount` |
| POST  | /api/boq | add `{projectId, itemNo?, description, unit?, quantity, unitRate}` (not viewer) |
| PATCH | /api/boq/{id} | update item fields (not viewer) |
| DELETE| /api/boq/{id} | remove item (not viewer) |

## Site updates

| Method | Path | Description |
|--------|------|-------------|
| GET  | /api/site-updates?projectId | list with materials used |
| POST | /api/site-updates | submit daily report (not viewer) |

## Reports and intelligence

| Method | Path | Description |
|--------|------|-------------|
| GET  | /api/dashboard | stats, projects, alerts, deadlines, activity |
| GET  | /api/activity?projectId | activity feed |
| POST | /api/assistant | `{question, projectId?}` → computed answer from live data |

## Newsletter

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/newsletter/subscribe | `{email}` → `{subscribed}` (public; `409` when already subscribed) |

## Workspace

| Method | Path | Description |
|--------|------|-------------|
| GET  | /api/team | members + assignments |
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
