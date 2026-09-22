ALTER TABLE projects
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS expected_end_date,
    DROP COLUMN IF EXISTS start_date,
    DROP COLUMN IF EXISTS client_name;
