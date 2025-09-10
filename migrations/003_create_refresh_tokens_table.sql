-- Migration to create refresh_tokens table
-- This file is for reference only - GORM AutoMigrate handles schema creation

CREATE TABLE IF NOT EXISTS `refresh_tokens` (
    `id` char(36) NOT NULL,
    `user_id` char(36) NOT NULL,
    `tenant_id` char(36) NOT NULL,
    `token` text NOT NULL,
    `expires_at` datetime(3) NOT NULL,
    `is_revoked` boolean DEFAULT FALSE,
    `created_at` datetime(3) NULL,
    `updated_at` datetime(3) NULL,
    `deleted_at` datetime(3) NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_refresh_tokens_token` (`token`(255)),
    KEY `idx_refresh_tokens_deleted_at` (`deleted_at`),
    KEY `idx_refresh_tokens_user_id` (`user_id`),
    KEY `idx_refresh_tokens_tenant_id` (`tenant_id`),
    KEY `idx_refresh_tokens_expires_at` (`expires_at`),
    CONSTRAINT `fk_refresh_tokens_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_refresh_tokens_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;