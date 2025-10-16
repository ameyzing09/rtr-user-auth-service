-- Create audit_logs table for security event logging
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    event_id CHAR(36) NOT NULL UNIQUE,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    action VARCHAR(100) NOT NULL,
    actor_id CHAR(36),
    actor_tenant_id CHAR(36),
    actor_role VARCHAR(50),
    target_resource_id VARCHAR(255),
    target_resource_type VARCHAR(50),
    target_tenant_id CHAR(36),
    status ENUM('success', 'denied', 'error') NOT NULL,
    reason VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,
    metadata JSON,
    INDEX idx_timestamp (timestamp),
    INDEX idx_action (action),
    INDEX idx_actor_id (actor_id),
    INDEX idx_target_tenant (target_tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
