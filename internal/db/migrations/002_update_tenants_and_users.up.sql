-- Update tenants: add domain and plan
ALTER TABLE tenants
  ADD COLUMN domain VARCHAR(255) NOT NULL UNIQUE AFTER name,
  ADD COLUMN plan VARCHAR(50) NOT NULL DEFAULT 'FREE' AFTER domain;

-- Update users: add force_password_reset
ALTER TABLE users
  ADD COLUMN force_password_reset BOOLEAN NOT NULL DEFAULT FALSE AFTER role;
