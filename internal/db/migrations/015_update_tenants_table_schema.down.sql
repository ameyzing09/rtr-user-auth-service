  -- Drop new index
  DROP INDEX IF EXISTS `idx_tenant_slug` ON `tenants`;

  -- Remove unique constraint
  DROP INDEX IF EXISTS `IDX_2310ecc5cb8be427097154b18f` ON `tenants`;

  -- Revert slug column
  ALTER TABLE `tenants` MODIFY COLUMN `slug` VARCHAR(50) NULL;

  -- Revert ID column
  ALTER TABLE `tenants` MODIFY COLUMN `id` CHAR(36) NOT NULL;

  -- Restore old columns
  ALTER TABLE `tenants` ADD COLUMN `deleted_at` TIMESTAMP NULL;
  ALTER TABLE `tenants` ADD COLUMN `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;
  ALTER TABLE `tenants` ADD COLUMN `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
  ALTER TABLE `tenants` ADD COLUMN `failed_reason` TEXT NULL;
  ALTER TABLE `tenants` ADD COLUMN `created_by` CHAR(36) NULL;
  ALTER TABLE `tenants` ADD COLUMN `status` ENUM('PENDING', 'PROVISIONING', 'AWAITING_BRANDING', 'ACTIVE', 'FAILED', 'SUSPENDED',
   'DELETED') NOT NULL DEFAULT 'PENDING';
  ALTER TABLE `tenants` ADD COLUMN `plan` ENUM('BASIC', 'STARTER', 'GROWTH', 'ENTERPRISE', 'ON_PREM') NULL;
  ALTER TABLE `tenants` ADD COLUMN `domain` VARCHAR(255) NULL;

  -- Restore old indexes
  CREATE UNIQUE INDEX `ux_tenants_slug` ON `tenants` (`slug`);
  CREATE UNIQUE INDEX `ux_tenants_name` ON `tenants` (`name`);
  CREATE INDEX `idx_tenants_status` ON `tenants` (`status`);
  CREATE INDEX `idx_tenants_plan` ON `tenants` (`plan`);
  CREATE UNIQUE INDEX `domain` ON `tenants` (`domain`);