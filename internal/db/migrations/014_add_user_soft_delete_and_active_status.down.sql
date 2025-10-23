-- Revert soft delete and active status support from users table

-- Drop user archives table
DROP TABLE IF EXISTS user_archives;

-- Drop index on deleted_at
DROP INDEX IF EXISTS idx_users_deleted_at ON users;

-- Remove soft delete and active status columns
ALTER TABLE users DROP COLUMN deleted_at;
ALTER TABLE users DROP COLUMN is_active;
