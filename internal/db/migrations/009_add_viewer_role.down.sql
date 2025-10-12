-- Remove VIEWER role from users.role ENUM
-- Note: This will fail if any users have the VIEWER role
ALTER TABLE users MODIFY COLUMN role ENUM('SUPERADMIN','ADMIN','HR','INTERVIEWER','CANDIDATE') NOT NULL DEFAULT 'CANDIDATE';
