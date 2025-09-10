-- Migration to create tenants table
-- This file is for reference only - GORM AutoMigrate handles schema creation

CREATE TABLE IF NOT EXISTS `tenants` (
    `id` char(36) NOT NULL,
    `name` varchar(100) NOT NULL,
    `domain` varchar(255) NOT NULL,
    `is_active` boolean DEFAULT TRUE,
    `created_at` datetime(3) NULL,
    `updated_at` datetime(3) NULL,
    `deleted_at` datetime(3) NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_tenants_name` (`name`),
    UNIQUE KEY `idx_tenants_domain` (`domain`),
    KEY `idx_tenants_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;