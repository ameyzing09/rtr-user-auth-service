-- Remove tenant archives table and soft delete support
DROP TABLE IF EXISTS tenant_archives;
ALTER TABLE tenants DROP COLUMN deleted_at;