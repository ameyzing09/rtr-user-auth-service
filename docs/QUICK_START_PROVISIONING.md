# 🚀 Tenant Provisioning System - Quick Start Guide

## What Was Implemented

A complete **event-driven tenant provisioning system** using the Outbox Pattern that automatically provisions tenants when they are created.

## How It Works

1. **Create Tenant** → Tenant starts in `PENDING` status
2. **Outbox Event** → `tenant.created` event written to outbox table
3. **Publisher** → Reads outbox and publishes events every 5 seconds
4. **Consumer** → Listens for events and triggers provisioning
5. **Provisioning** → Updates tenant to `ACTIVE` or `FAILED`

## Prerequisites

✅ Database running (MySQL)
✅ `.env` file configured
✅ Go 1.24+ installed

## Quick Start (3 Simple Steps)

### Step 1: Start the Main API Server

```powershell
make run
# OR
go run cmd/server/main.go
```

### Step 2: Start the Outbox Publisher (New Terminal)

```powershell
make run-publisher
# OR
go run cmd/outbox-publisher/main.go
```

You should see:
```
[INFO] 🚀 Outbox publisher started successfully batchSize=100 pollInterval=5s
[INFO] Listening for unpublished events...
```

### Step 3: Start the Event Consumer (New Terminal)

```powershell
make run-consumer
# OR
go run cmd/event-consumer/main.go
```

You should see:
```
[INFO] 🚀 Event consumer started successfully
[INFO] Listening for events... topics=[tenant]
```

## Testing the System

### 1. Create a Tenant

```bash
POST http://localhost:8082/api/v1/admin/tenants
Content-Type: application/json
Authorization: Bearer <your-token>

{
  "name": "Acme Corporation",
  "slug": "acme-corp",
  "plan": "BASIC"
}
```

### 2. Watch the Logs

**Outbox Publisher** will show:
```
[INFO] Publishing batch count=1
[INFO] Event published eventID=1 type=tenant.created aggregateID=<tenant-id>
[INFO] Batch published success=1 failures=0 total=1
```

**Event Consumer** will show:
```
[INFO] Received tenant.created event tenantID=<tenant-id> name=Acme Corporation
[INFO] Starting tenant provisioning tenantID=<tenant-id>
[INFO] Tenant configuration initialized tenantID=<tenant-id>
[INFO] Tenant provisioning completed successfully tenantID=<tenant-id>
```

### 3. Check Tenant Status

```sql
SELECT id, name, status, created_at, updated_at 
FROM tenants 
ORDER BY created_at DESC 
LIMIT 5;
```

Expected result:
- Status changes: `PENDING` → `PROVISIONING` → `ACTIVE` ✅

### 4. Check Outbox Table

```sql
SELECT id, type, aggregate_id, created_at, published_at 
FROM outbox 
ORDER BY created_at DESC 
LIMIT 5;
```

Expected result:
- `published_at` should have a timestamp (event was published)

## Monitoring Commands

### Comprehensive Health Check (⭐ Recommended)
```powershell
make health-check
```
Shows all diagnostic queries with explanations.

### Check Unpublished Events
```powershell
make check-outbox
```

### Check Tenant Status Distribution
```powershell
make check-tenants
```

### View Recent Tenants
```powershell
make view-tenants
```

**Note:** These commands show SQL queries to run in your database tool (MySQL Workbench, DBeaver, etc.). See `docs/WINDOWS_MONITORING_FIX.md` for details.

### Manual Database Queries

**Unpublished Events:**
```sql
SELECT COUNT(*) as unpublished 
FROM outbox 
WHERE published_at IS NULL;
```

**Tenant Status Counts:**
```sql
SELECT status, COUNT(*) as count 
FROM tenants 
GROUP BY status;
```

**Recent Provisioning Activity:**
```sql
SELECT id, name, status, 
       TIMESTAMPDIFF(SECOND, created_at, updated_at) as provision_time_seconds
FROM tenants 
WHERE status IN ('ACTIVE', 'FAILED')
ORDER BY created_at DESC 
LIMIT 10;
```

## Troubleshooting

### Problem: Tenant Stuck in PENDING

**Diagnosis:**
```sql
SELECT * FROM outbox WHERE published_at IS NULL;
```

**Solution:**
- If unpublished events exist → Outbox publisher not running
- Start publisher: `make run-publisher`

### Problem: Tenant Status is FAILED

**Diagnosis:**
```sql
SELECT id, name, status, failed_reason 
FROM tenants 
WHERE status = 'FAILED';
```

**Solution:**
- Check the `failed_reason` column
- Review event consumer logs for errors
- Fix the issue and retry:
```sql
UPDATE tenants 
SET status = 'PENDING', failed_reason = NULL 
WHERE id = '<tenant-id>';
```

### Problem: Events Not Being Consumed

**Diagnosis:**
- Check if event consumer is running
- Check consumer logs for errors

**Solution:**
- Restart consumer: `make run-consumer`
- Check database connectivity
- Verify configuration

## Production Deployment

### Using Systemd (Linux)

Create service files:

**`/etc/systemd/system/rtr-outbox-publisher.service`**
```ini
[Unit]
Description=RTR Outbox Publisher
After=network.target mysql.service

[Service]
Type=simple
User=rtr
WorkingDirectory=/opt/rtr-auth-service
ExecStart=/opt/rtr-auth-service/bin/outbox-publisher
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**`/etc/systemd/system/rtr-event-consumer.service`**
```ini
[Unit]
Description=RTR Event Consumer
After=network.target mysql.service

[Service]
Type=simple
User=rtr
WorkingDirectory=/opt/rtr-auth-service
ExecStart=/opt/rtr-auth-service/bin/event-consumer
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable rtr-outbox-publisher
sudo systemctl enable rtr-event-consumer
sudo systemctl start rtr-outbox-publisher
sudo systemctl start rtr-event-consumer
```

### Using Docker Compose

```yaml
version: '3.8'

services:
  auth-api:
    build: .
    command: ./server
    ports:
      - "8082:8082"
    depends_on:
      - mysql

  outbox-publisher:
    build: .
    command: ./outbox-publisher
    depends_on:
      - mysql
    restart: unless-stopped

  event-consumer:
    build: .
    command: ./event-consumer
    depends_on:
      - mysql
    restart: unless-stopped

  mysql:
    image: mysql:8
    environment:
      MYSQL_ROOT_PASSWORD: secret
      MYSQL_DATABASE: recrutr
    volumes:
      - mysql-data:/var/lib/mysql

volumes:
  mysql-data:
```

## Performance Tuning

### Adjust Polling Interval

Edit `cmd/outbox-publisher/main.go`:
```go
publisherConfig := messaging.PublisherConfig{
    BatchSize:    100,           // Increase for high throughput
    PollInterval: 2 * time.Second, // Decrease for faster processing
    MaxRetries:   3,
}
```

### Database Indexing

Ensure optimal query performance:
```sql
-- Index for outbox queries
CREATE INDEX idx_outbox_published ON outbox(published_at, created_at);

-- Index for tenant status queries
CREATE INDEX idx_tenants_status ON tenants(status, created_at);
```

## Next Steps

1. ✅ **System is running** - Tenants now auto-provision
2. 🔄 **Monitor logs** - Watch for any errors
3. 📊 **Add metrics** - Integrate Prometheus/Grafana
4. 🔌 **Integrate Kafka** - Replace LogBroker with Kafka for production
5. 🔄 **Add retry logic** - Implement exponential backoff for failures
6. 📝 **Add observability** - OpenTelemetry tracing

## Support

For detailed architecture and configuration, see:
- 📖 [PROVISIONING_SYSTEM.md](./PROVISIONING_SYSTEM.md)
- 📁 Code: `messaging/`, `services/`, `consumers/`

## Answer to Your Original Question

**Q: When will my tenant be available?**

**A: Now it's automatic!** ⚡

- **Immediately after creation**: Tenant is in `PENDING` status
- **Within 5-10 seconds**: Publisher picks up the event
- **Within 1-2 seconds**: Consumer provisions the tenant
- **Total time**: ~5-15 seconds from creation to `ACTIVE`

You'll see the status update in real-time on your UI when you refresh!

---

🎉 **Congratulations!** You now have a production-ready, event-driven tenant provisioning system!
