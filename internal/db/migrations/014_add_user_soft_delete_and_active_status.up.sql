-- Add soft delete and active status support to users table
ALTER TABLE users ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;

-- Add index on deleted_at for performance
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Create user archives table for compliance and audit purposes
CREATE TABLE user_archives (
    id CHAR(36) PRIMARY KEY,
    tenant_id CHAR(36) NOT NULL,
    name VARCHAR(150) NOT NULL,
    email VARCHAR(190) NOT NULL,
    role ENUM('SUPERADMIN','ADMIN','HR','INTERVIEWER','VIEWER','CANDIDATE') NOT NULL,
    is_owner BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_by CHAR(36),
    deleted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    delete_reason TEXT,
    INDEX idx_user_archives_deleted_at (deleted_at),
    INDEX idx_user_archives_deleted_by (deleted_by),
    INDEX idx_user_archives_tenant_id (tenant_id),
    INDEX idx_user_archives_email (email)
);
