-- The projects repository (backend/internal/project/postgres_repo.go) selects
-- and writes client_name, start_date, expected_end_date and updated_at, but no
-- earlier migration created them. This migration reconciles the schema with
-- the code so project CRUD stops failing at runtime.

ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS client_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS start_date DATE,
    ADD COLUMN IF NOT EXISTS expected_end_date DATE,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP;
