# 🚀 Provisioning System - Cheat Sheet

## Quick Start (Copy & Paste)

### Start All Services
```powershell
# Terminal 1
make run

# Terminal 2  
make run-publisher

# Terminal 3
make run-consumer
```

## Status Flow
```
PENDING → PROVISIONING → ACTIVE (5-15 seconds)
                ↓
              FAILED (if error)
```

## Key Commands
```powershell
make build-all          # Build everything
make check-outbox       # Check unpublished events
make check-tenants      # Check status distribution
make test-provisioning  # Run tests
```

## Troubleshooting One-Liners

```sql
-- Stuck in PENDING?
SELECT * FROM outbox WHERE published_at IS NULL;

-- Check provisioning time
SELECT TIMESTAMPDIFF(SECOND, created_at, updated_at) 
FROM tenants WHERE status='ACTIVE' ORDER BY created_at DESC LIMIT 1;

-- Retry failed tenant
UPDATE tenants SET status='PENDING', failed_reason=NULL WHERE id='<id>';
```

## Logs to Watch

**Good:**
```
[INFO] Event published
[INFO] Tenant provisioning completed successfully
```

**Bad:**
```
[ERROR] Failed to publish event
[ERROR] Provisioning failed
```

## Files You Need to Know

- `cmd/outbox-publisher/main.go` - Publisher
- `cmd/event-consumer/main.go` - Consumer
- `services/tenant_provisioning.go` - Business logic
- `messaging/outbox_publisher.go` - Core publisher
- `docs/QUICK_START_PROVISIONING.md` - Full guide

## Answer: When Will My Tenant Be Available?

**5-15 seconds, automatically!** 🎉

No manual action needed. Just wait and refresh.
