ALTER TABLE users
  DROP COLUMN is_owner;

ALTER TABLE tenants
  DROP INDEX idx_tenants_plan,
  DROP INDEX idx_tenants_status,
  DROP INDEX ux_tenants_slug;

ALTER TABLE tenants
  DROP COLUMN failed_reason,
  DROP COLUMN created_by,
  DROP COLUMN status,
  DROP COLUMN slug,
  MODIFY COLUMN plan VARCHAR(50) NOT NULL DEFAULT 'FREE' AFTER domain,
  MODIFY COLUMN domain VARCHAR(255) NOT NULL;

UPDATE tenants
SET plan = 'FREE'
WHERE plan IS NULL;
