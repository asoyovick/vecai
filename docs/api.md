# VECAI API Specification

## Purpose

This document defines the HTTP API exposed by the VECAI backend.

The backend is a Go/Gin JSON API. The Next.js frontend and future clients communicate with it over HTTP.

---

## Base URL

Development default:

```text
http://localhost:8080/api/v1
```

Health check (outside `/api/v1`):

```text
GET /health
```

---

## API Principles

* JSON request and response bodies
* Resource-oriented routes under `/api/v1`
* JWT bearer authentication for protected operations
* Server-side validation and authorization
* Consistent response envelope

---

## Authentication

Authentication uses JWTs. Clients obtain a token from auth endpoints and send it as:

```text
Authorization: Bearer <token>
```

The API also uses optional auth middleware on the `/api/v1` group; individual handlers enforce authentication/authorization as needed.

---

## Standard Response Format

Success:

```json
{
  "success": true,
  "message": "Project created successfully",
  "data": {}
}
```

Error:

```json
{
  "success": false,
  "error": "Project not found"
}
```

Some handlers may include `message`; clients should handle `success`, `message`, `data`, and `error`.

---

## Implemented Endpoints

### Health

| Method | Path | Description |
| --- | --- | --- |
| GET | `/health` | Service health check |

### Auth

| Method | Path | Description |
| --- | --- | --- |
| POST | `/api/v1/auth/register` | Register a user |
| POST | `/api/v1/auth/login` | Login and receive auth token/session response |
| POST | `/api/v1/auth/verify` | Verify account/token flow |
| POST | `/api/v1/auth/logout` | Logout (client-side token discard; JWT is stateless) |

### Current User

| Method | Path | Description |
| --- | --- | --- |
| GET | `/api/v1/users/me` | Get current authenticated user |
| PUT | `/api/v1/users/me` | Update current user profile |

### Projects

| Method | Path | Description |
| --- | --- | --- |
| POST | `/api/v1/projects` | Create project |
| GET | `/api/v1/projects` | List projects |
| GET | `/api/v1/projects/:id` | Get project by ID |
| PUT | `/api/v1/projects/:id` | Update project |
| DELETE | `/api/v1/projects/:id` | Delete project |

### Project Members

| Method | Path | Description |
| --- | --- | --- |
| POST | `/api/v1/projects/:id/members` | Invite/add project member |
| GET | `/api/v1/projects/:id/members` | List project members |
| DELETE | `/api/v1/projects/:id/members/:memberId` | Remove project member |

### Project Blueprints

| Method | Path | Description |
| --- | --- | --- |
| POST | `/api/v1/projects/:id/blueprints` | Add blueprint metadata |
| GET | `/api/v1/projects/:id/blueprints` | List blueprints |

### Project Specs

| Method | Path | Description |
| --- | --- | --- |
| POST | `/api/v1/projects/:id/specs` | Add spec |
| GET | `/api/v1/projects/:id/specs` | List specs |

### Materials

| Method | Path | Description |
| --- | --- | --- |
| POST | `/api/v1/projects/:id/materials` | Record material delivery |
| GET | `/api/v1/projects/:id/materials` | List materials |
| GET | `/api/v1/projects/:id/materials/summary` | Material summary |
| PUT | `/api/v1/projects/:id/materials/:materialId` | Update material record |
| DELETE | `/api/v1/projects/:id/materials/:materialId` | Delete material record |

### Daily Reports

| Method | Path | Description |
| --- | --- | --- |
| POST | `/api/v1/projects/:id/reports` | Create daily report |
| GET | `/api/v1/projects/:id/reports` | List daily reports |
| GET | `/api/v1/projects/:id/reports/summary` | Reports summary |
| GET | `/api/v1/projects/:id/reports/:reportId` | Get daily report |
| PUT | `/api/v1/projects/:id/reports/:reportId` | Update daily report |
| DELETE | `/api/v1/projects/:id/reports/:reportId` | Delete daily report |

### BOQ

| Method | Path | Description |
| --- | --- | --- |
| POST | `/api/v1/projects/:id/boq` | Create BOQ item |
| GET | `/api/v1/projects/:id/boq` | List BOQ items |
| GET | `/api/v1/projects/:id/boq/compare` | Compare estimated vs actual BOQ costs |
| PUT | `/api/v1/projects/:id/boq/:itemId` | Update BOQ item |
| DELETE | `/api/v1/projects/:id/boq/:itemId` | Delete BOQ item |

### Documents

| Method | Path | Description |
| --- | --- | --- |
| POST | `/api/v1/projects/:id/documents` | Upload project document (multipart/form-data) |
| GET | `/api/v1/projects/:id/documents` | List project documents |
| GET | `/api/v1/projects/:id/documents/:documentId` | Download document |
| DELETE | `/api/v1/projects/:id/documents/:documentId` | Delete document |

### Newsletter

| Method | Path | Description |
| --- | --- | --- |
| POST | `/api/v1/newsletter/subscribe` | Subscribe email to newsletter |

---

## Planned / Not Yet Implemented

The following endpoints/features appear in earlier design docs or roadmap but are not implemented in the current API:

* Suppliers CRUD (`/suppliers`)
* Quotations and quotation comparison (`/projects/:id/quotations`)
* Project dashboard aggregate endpoint (`/dashboard/projects/:id`)
* Notifications, timeline/activity, search
* AI endpoints (`/projects/:id/ai/*`)

---

## HTTP Status Codes

| Code | Meaning |
| --- | --- |
| 200 | OK |
| 201 | Created |
| 204 | No Content / preflight response |
| 400 | Bad Request / invalid input |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict (e.g. duplicate unique value) |
| 500 | Internal Server Error |

---

## CORS

CORS is configured by `CORS_ALLOWED_ORIGINS` (comma-separated). Development defaults include:

```text
http://localhost:3000,http://localhost:8080
```

---

## Security Notes

* Use HTTPS in production.
* Keep JWT secrets configured via environment (`JWT_SECRET`).
* Do not expose database errors or internal details to clients.
* Validate and authorize every project-scoped request.
