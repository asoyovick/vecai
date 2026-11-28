# VECAI Database Design

## Purpose

This document describes the current PostgreSQL schema used by VECAI, based on migrations in `backend/db/migrations/`.

The database stores persistent application data. File contents for documents are stored outside PostgreSQL; document metadata is stored in the database.

---

## Database Goals

* Maintain data integrity
* Support project-scoped queries
* Keep schema changes traceable through migrations
* Provide a foundation for future analytics and AI features

---

## Migration Overview

Current migrations (`000001` to `000011`):

| Migration | Purpose |
| --- | --- |
| 000001 | Create `projects` table |
| 000002 | Create `blueprints` and `specs` tables |
| 000003 | Create `users` table |
| 000004 | Create `project_members` table |
| 000005 | Create `materials` table |
| 000006 | Create `daily_reports` table |
| 000007 | Create `boq_items` table |
| 000008 | Create `project_documents` table |
| 000009 | Create `newsletter_subscribers` table |
| 000010 | Add profile/audit columns to `users` |
| 000011 | Add missing project columns used by repository code |

---

## Current Tables

### users

Stores user accounts.

| Column | Type | Notes |
| --- | --- | --- |
| id | UUID | Primary key, default `gen_random_uuid()` |
| name | VARCHAR(255) | Required |
| email | VARCHAR(255) | Required, unique |
| password_hash | TEXT | Required |
| role | VARCHAR(50) | Default `owner` |
| phone | VARCHAR(50) | Added in 000010 |
| is_verified | BOOLEAN | Default `TRUE`, added in 000010 |
| created_at | TIMESTAMPTZ | Default `NOW()` |
| updated_at | TIMESTAMPTZ | Default `CURRENT_TIMESTAMP`, added in 000010 |

### projects

Stores construction projects.

| Column | Type | Notes |
| --- | --- | --- |
| id | UUID | Primary key |
| name | VARCHAR(255) | Required |
| description | TEXT | Optional |
| location | VARCHAR(255) | Required |
| owner_id | VARCHAR(255) | Required (creator/owner identifier) |
| status | VARCHAR(50) | Default `Planning` |
| client_name | VARCHAR(255) | Added in 000011 |
| start_date | DATE | Added in 000011 |
| expected_end_date | DATE | Added in 000011 |
| created_at | TIMESTAMPTZ | Default `CURRENT_TIMESTAMP` |
| updated_at | TIMESTAMPTZ | Default `CURRENT_TIMESTAMP`, added in 000011 |

### project_members

Associates users with projects.

| Column | Type | Notes |
| --- | --- | --- |
| id | UUID | Primary key |
| project_id | UUID | FK to `projects(id)` ON DELETE CASCADE |
| user_id | UUID | FK to `users(id)` ON DELETE CASCADE |
| role | VARCHAR(50) | Required |
| joined_at | TIMESTAMPTZ | Default `CURRENT_TIMESTAMP` |

Constraints/indexes:

* Unique `(project_id, user_id)`
* Indexes on `project_id`, `user_id`

### blueprints

Stores blueprint metadata for a project.

| Column | Type | Notes |
| --- | --- | --- |
| id | UUID | Primary key |
| project_id | UUID | FK to `projects(id)` ON DELETE CASCADE |
| title | VARCHAR(255) | Required |
| description | TEXT | Optional |
| file_path | TEXT | Required |
| version | INT | Default 1 |
| uploader_id | VARCHAR(255) | Required |
| created_at | TIMESTAMPTZ | Default `CURRENT_TIMESTAMP` |
| updated_at | TIMESTAMPTZ | Default `CURRENT_TIMESTAMP` |

### specs

Stores project specifications.

| Column | Type | Notes |
| --- | --- | --- |
| id | UUID | Primary key |
| project_id | UUID | FK to `projects(id)` ON DELETE CASCADE |
| title | VARCHAR(255) | Required |
| description | TEXT | Required |
| section_code | VARCHAR(50) | Required |
| created_by | VARCHAR(255) | Required |
| created_at | TIMESTAMPTZ | Default `CURRENT_TIMESTAMP` |
| updated_at | TIMESTAMPTZ | Default `CURRENT_TIMESTAMP` |

### materials

Tracks material deliveries/records.

| Column | Type | Notes |
| --- | --- | --- |
| id | UUID | Primary key |
| project_id | UUID | FK to `projects(id)` ON DELETE CASCADE |
| material_name | VARCHAR(255) | Required |
| quantity | DECIMAL(14,3) | Required, `> 0` |
| unit | VARCHAR(50) | Required |
| unit_cost | DECIMAL(14,2) | Default 0, `>= 0` |
| supplier | VARCHAR(255) | Optional |
| delivery_date | DATE | Optional |
| recorded_by | VARCHAR(255) | Optional |
| created_at | TIMESTAMPTZ | Default `CURRENT_TIMESTAMP` |

Indexes: `project_id`, `delivery_date`.

### daily_reports

Stores daily site reports.

| Column | Type | Notes |
| --- | --- | --- |
| id | UUID | Primary key |
| project_id | UUID | FK to `projects(id)` ON DELETE CASCADE |
| user_id | VARCHAR(255) | Required |
| report_date | DATE | Required |
| completed_work | TEXT | Required |
| workers_present | INTEGER | Default 0, `>= 0` |
| remarks | TEXT | Optional |
| created_at | TIMESTAMPTZ | Default `CURRENT_TIMESTAMP` |

Indexes: `project_id`, `report_date`.

### boq_items

Stores Bill of Quantities line items.

| Column | Type | Notes |
| --- | --- | --- |
| id | UUID | Primary key |
| project_id | UUID | FK to `projects(id)` ON DELETE CASCADE |
| item_name | VARCHAR(255) | Required |
| quantity | DECIMAL(14,3) | Required, `> 0` |
| unit | VARCHAR(50) | Optional |
| estimated_cost | DECIMAL(14,2) | Default 0, `>= 0` |
| actual_cost | DECIMAL(14,2) | Default 0, `>= 0` |
| created_at | TIMESTAMPTZ | Default `CURRENT_TIMESTAMP` |
| updated_at | TIMESTAMPTZ | Default `CURRENT_TIMESTAMP` |

Index: `project_id`.

### project_documents

Stores document metadata; files live in storage.

| Column | Type | Notes |
| --- | --- | --- |
| id | UUID | Primary key |
| project_id | UUID | FK to `projects(id)` ON DELETE CASCADE |
| uploaded_by | VARCHAR(255) | Required |
| filename | VARCHAR(255) | Required |
| file_type | VARCHAR(100) | Optional MIME/type label |
| file_size | BIGINT | Default 0 |
| storage_path | TEXT | Required storage key/path |
| category | VARCHAR(50) | Optional |
| uploaded_at | TIMESTAMPTZ | Default `CURRENT_TIMESTAMP` |

Index: `project_id`.

### newsletter_subscribers

Stores newsletter signups.

| Column | Type | Notes |
| --- | --- | --- |
| id | UUID | Primary key |
| email | VARCHAR(255) | Required, unique |
| subscribed_at | TIMESTAMPTZ | Default `CURRENT_TIMESTAMP` |

---

## Relationships

* One project has many blueprints, specs, members, materials, daily reports, BOQ items, and documents.
* Users and projects have a many-to-many relationship through `project_members`.
* Most child records are removed when a project is deleted via `ON DELETE CASCADE`.

---

## Not Yet Implemented as Tables

The earlier design discussed tables that are not present in current migrations:

* Suppliers
* Quotations
* Notifications
* Payments/equipment/inventory/AI prediction tables

Add these through new migrations when the corresponding features are implemented.

---

## Schema Notes

* Some identifier columns (`owner_id`, `uploader_id`, `created_by`, `recorded_by`, `user_id` in reports/documents) are stored as `VARCHAR(255)` rather than UUID foreign keys. Future migrations may normalize these where appropriate.
* Timestamps are included on most major tables; audit fields can be expanded as needed.
* Prefer additive migrations and avoid editing applied migrations.
