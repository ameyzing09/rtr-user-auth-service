-- Add soft delete support to tenants table and create tenant archives table
ALTER TABLE tenants ADD COLUMN deleted_at TIMESTAMP NULL;

-- Create tenant archives table
CREATE TABLE tenant_archives (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255),
    slug VARCHAR(50),
    plan ENUM('BASIC','STARTER','GROWTH','ENTERPRISE','ON_PREM'),
    status ENUM('PENDING','PROVISIONING','AWAITING_BRANDING','ACTIVE','FAILED','SUSPENDED','DELETED') NOT NULL,
    created_by CHAR(36),
    failed_reason TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_by CHAR(36) NOT NULL,
    deleted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    delete_reason TEXT,
    INDEX idx_tenant_archives_deleted_at (deleted_at),
    INDEX idx_tenant_archives_deleted_by (deleted_by)
);