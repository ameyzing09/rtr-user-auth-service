-- If any SUPERADMIN exists, demote to ADMIN before shrinking enum
UPDATE users SET role = 'ADMIN' WHERE role = 'SUPERADMIN';

ALTER TABLE users
  MODIFY COLUMN role ENUM('ADMIN','HR','INTERVIEWER','CANDIDATE')
  NOT NULL DEFAULT 'CANDIDATE';
