-- Migration 010: Convert all TIMESTAMP columns from IST (+05:30) to UTC (+00:00)
-- This migration converts existing timestamp data to UTC to match the loc=UTC DSN parameter

-- Convert tenants table timestamps
UPDATE tenants
SET
  created_at = CONVERT_TZ(created_at, '+05:30', '+00:00'),
  updated_at = CONVERT_TZ(updated_at, '+05:30', '+00:00')
WHERE created_at IS NOT NULL;

-- Convert users table timestamps
UPDATE users
SET
  created_at = CONVERT_TZ(created_at, '+05:30', '+00:00'),
  updated_at = CONVERT_TZ(updated_at, '+05:30', '+00:00')
WHERE created_at IS NOT NULL;

-- Convert tenant_archives table timestamps
UPDATE tenant_archives
SET
  created_at = CONVERT_TZ(created_at, '+05:30', '+00:00'),
  updated_at = CONVERT_TZ(updated_at, '+05:30', '+00:00'),
  deleted_at = CONVERT_TZ(deleted_at, '+05:30', '+00:00')
WHERE created_at IS NOT NULL;

-- Convert subscriptions table timestamps (including nullable columns)
UPDATE subscriptions
SET
  created_at = CONVERT_TZ(created_at, '+05:30', '+00:00'),
  updated_at = CONVERT_TZ(updated_at, '+05:30', '+00:00'),
  period_start = CASE WHEN period_start IS NOT NULL THEN CONVERT_TZ(period_start, '+05:30', '+00:00') ELSE NULL END,
  period_end = CASE WHEN period_end IS NOT NULL THEN CONVERT_TZ(period_end, '+05:30', '+00:00') ELSE NULL END,
  trial_ends_at = CASE WHEN trial_ends_at IS NOT NULL THEN CONVERT_TZ(trial_ends_at, '+05:30', '+00:00') ELSE NULL END,
  next_renewal_at = CASE WHEN next_renewal_at IS NOT NULL THEN CONVERT_TZ(next_renewal_at, '+05:30', '+00:00') ELSE NULL END,
  canceled_at = CASE WHEN canceled_at IS NOT NULL THEN CONVERT_TZ(canceled_at, '+05:30', '+00:00') ELSE NULL END
WHERE created_at IS NOT NULL;

-- Convert outbox table timestamps
UPDATE outbox
SET
  created_at = CONVERT_TZ(created_at, '+05:30', '+00:00'),
  published_at = CASE WHEN published_at IS NOT NULL THEN CONVERT_TZ(published_at, '+05:30', '+00:00') ELSE NULL END
WHERE created_at IS NOT NULL;

-- Convert idempotency_keys table timestamps
UPDATE idempotency_keys
SET
  created_at = CONVERT_TZ(created_at, '+05:30', '+00:00')
WHERE created_at IS NOT NULL;
