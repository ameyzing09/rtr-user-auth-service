-- Rollback for 001_create_tenants_and_users.sql

-- Drop child table first to satisfy FK
DROP TABLE IF EXISTS users;

-- Then drop parent table
DROP TABLE IF EXISTS tenants;
