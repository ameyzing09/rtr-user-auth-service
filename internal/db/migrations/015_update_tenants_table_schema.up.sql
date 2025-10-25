-- Drop old indexes
  DROP INDEX IF EXISTS `domain` ON `tenants`;
  DROP INDEX IF EXISTS `idx_tenants_plan` ON `tenants`;
  DROP INDEX IF EXISTS `idx_tenants_status` ON `tenants`;
  DROP INDEX IF EXISTS `ux_tenants_name` ON `tenants`;
  DROP INDEX IF EXISTS `ux_tenants_slug` ON `tenants`;

  -- Drop old columns
  ALTER TABLE `tenants` DROP COLUMN IF EXISTS `domain`;
  ALTER TABLE `tenants` DROP COLUMN IF EXISTS `plan`;
  ALTER TABLE `tenants` DROP COLUMN IF EXISTS `status`;
  ALTER TABLE `tenants` DROP COLUMN IF EXISTS `created_by`;
  ALTER TABLE `tenants` DROP COLUMN IF EXISTS `failed_reason`;
  ALTER TABLE `tenants` DROP COLUMN IF EXISTS `created_at`;
  ALTER TABLE `tenants` DROP COLUMN IF EXISTS `updated_at`;
  ALTER TABLE `tenants` DROP COLUMN IF EXISTS `deleted_at`;

  -- Update ID column type (if needed)
  -- Note: This may require careful handling of existing data
  ALTER TABLE `tenants` MODIFY COLUMN `id` VARCHAR(36) NOT NULL;

  -- Update slug column
  ALTER TABLE `tenants` MODIFY COLUMN `slug` VARCHAR(255) NOT NULL;

  -- Add unique constraint for slug
  ALTER TABLE `tenants` ADD UNIQUE INDEX `IDX_2310ecc5cb8be427097154b18f` (`slug`);

  -- Create new index on slug
  CREATE INDEX `idx_tenant_slug` ON `tenants` (`slug`);