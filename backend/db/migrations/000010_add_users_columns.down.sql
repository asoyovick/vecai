ALTER TABLE users
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS is_verified,
    DROP COLUMN IF EXISTS updated_at;
