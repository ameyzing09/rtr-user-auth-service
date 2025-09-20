ALTER TABLE users
  DROP COLUMN force_password_reset;

-- Remove columns added to tenants (domain has an implicit UNIQUE index that will be dropped with the column)
ALTER TABLE tenants
  DROP COLUMN plan,
  DROP COLUMN domain;