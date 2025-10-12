-- Rollback Migration 010: Convert timestamps back from UTC (+00:00) to IST (+05:30)
-- WARNING: Only use this if rolling back before the loc=UTC DSN change is deployed

-- Convert tenants table timestamps back to IST
UPDATE tenants
SET
  created_at = CONVERT_TZ(created_at, '+00:00', '+05:30'),
  updated_at = CONVERT_TZ(updated_at, '+00:00', '+05:30')
WHERE created_at IS NOT NULL;

-- Convert users table timestamps back to IST
UPDATE users
SET
  created_at = CONVERT_TZ(created_at, '+00:00', '+05:30'),
  updated_at = CONVERT_TZ(updated_at, '+00:00', '+05:30')
WHERE created_at IS NOT NULL;

-- Convert tenant_archives table timestamps back to IST
UPDATE tenant_archives
SET
  created_at = CONVERT_TZ(created_at, '+00:00', '+05:30'),
  updated_at = CONVERT_TZ(updated_at, '+00:00', '+05:30'),
  deleted_at = CONVERT_TZ(deleted_at, '+00:00', '+05:30')
WHERE created_at IS NOT NULL;

-- Convert subscriptions table timestamps back to IST (including nullable columns)
UPDATE subscriptions
SET
  created_at = CONVERT_TZ(created_at, '+00:00', '+05:30'),
  updated_at = CONVERT_TZ(updated_at, '+00:00', '+05:30'),
  period_start = CASE WHEN period_start IS NOT NULL THEN CONVERT_TZ(period_start, '+00:00', '+05:30') ELSE NULL END,
  period_end = CASE WHEN period_end IS NOT NULL THEN CONVERT_TZ(period_end, '+00:00', '+05:30') ELSE NULL END,
  trial_ends_at = CASE WHEN trial_ends_at IS NOT NULL THEN CONVERT_TZ(trial_ends_at, '+00:00', '+05:30') ELSE NULL END,
  next_renewal_at = CASE WHEN next_renewal_at IS NOT NULL THEN CONVERT_TZ(next_renewal_at, '+00:00', '+05:30') ELSE NULL END,
  canceled_at = CASE WHEN canceled_at IS NOT NULL THEN CONVERT_TZ(canceled_at, '+00:00', '+05:30') ELSE NULL END
WHERE created_at IS NOT NULL;

-- Convert outbox table timestamps back to IST
UPDATE outbox
SET
  created_at = CONVERT_TZ(created_at, '+00:00', '+05:30'),
  published_at = CASE WHEN published_at IS NOT NULL THEN CONVERT_TZ(published_at, '+00:00', '+05:30') ELSE NULL END
WHERE created_at IS NOT NULL;

-- Convert idempotency_keys table timestamps back to IST
UPDATE idempotency_keys
SET
  created_at = CONVERT_TZ(created_at, '+00:00', '+05:30')
WHERE created_at IS NOT NULL;
