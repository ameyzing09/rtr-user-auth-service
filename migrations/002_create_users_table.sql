-- Migration to create users table
-- This file is for reference only - GORM AutoMigrate handles schema creation

CREATE TABLE IF NOT EXISTS `users` (
    `id` char(36) NOT NULL,
    `tenant_id` char(36) NOT NULL,
    `email` varchar(255) NOT NULL,
    `password` varchar(255) NOT NULL,
    `first_name` varchar(50) NOT NULL,
    `last_name` varchar(50) NOT NULL,
    `role` enum('ADMIN','HR','INTERVIEWER','CANDIDATE') NOT NULL,
    `is_active` boolean DEFAULT TRUE,
    `created_at` datetime(3) NULL,
    `updated_at` datetime(3) NULL,
    `deleted_at` datetime(3) NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_tenant_email` (`tenant_id`,`email`),
    KEY `idx_users_deleted_at` (`deleted_at`),
    KEY `idx_users_tenant_id` (`tenant_id`),
    CONSTRAINT `fk_users_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;