CREATE TABLE subscriptions (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id       CHAR(36) NOT NULL,
  plan            ENUM('BASIC','STARTER','GROWTH','ENTERPRISE','ON_PREM') NOT NULL,
  billing_cycle   ENUM('MONTHLY','ANNUAL') NOT NULL DEFAULT 'MONTHLY',
  status          ENUM('TRIAL','ACTIVE','GRACE','SUSPENDED','CANCELED') NOT NULL DEFAULT 'TRIAL',
  currency        CHAR(3) NOT NULL DEFAULT 'USD',
  amount_cents    INT UNSIGNED NOT NULL DEFAULT 0,

  period_start    DATETIME NULL,
  period_end      DATETIME NULL,
  trial_ends_at   DATETIME NULL,
  next_renewal_at DATETIME NULL,
  canceled_at     DATETIME NULL,

  updated_by      CHAR(36) NULL,
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  UNIQUE KEY ux_subscription_tenant (tenant_id),
  KEY idx_subscription_status (status),
  CONSTRAINT fk_subscription_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id)
    ON UPDATE CASCADE ON DELETE CASCADE
);
