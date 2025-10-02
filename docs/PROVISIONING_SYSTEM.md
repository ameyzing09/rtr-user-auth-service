# Tenant Provisioning System

## Overview

This system implements the **Outbox Pattern** for reliable, asynchronous tenant provisioning in a multi-tenant architecture.

## Architecture

```
┌─────────────────┐
│  Auth Service   │ ──┐
│  (API)          │   │ 1. Create Tenant (PENDING status)
└─────────────────┘   │
                      ▼
                ┌──────────────┐
                │    Outbox    │
                │    Table     │
                └──────────────┘
                      │
                      │ 2. Outbox Publisher reads unpublished events
                      ▼
           ┌──────────────────────┐
           │  Outbox Publisher    │
           │  (Background Service)│
           └──────────────────────┘
                      │
                      │ 3. Publish to Message Bus
                      ▼
              ┌──────────────┐
              │ Message Broker│
              │  (Log/Kafka)  │
              └──────────────┘
                      │
                      │ 4. Consume tenant.created event
                      ▼
           ┌──────────────────────┐
           │  Event Consumer      │
           │  (Background Service)│
           └──────────────────────┘
                      │
                      │ 5. Provision Tenant
                      ▼
              ┌──────────────┐
              │ PROVISIONING │──► ACTIVE / FAILED
              └──────────────┘
```

## Components

### 1. **Outbox Table** (`models/outbox.go`)
Stores events to be published, ensuring at-least-once delivery.

### 2. **Outbox Publisher** (`cmd/outbox-publisher/main.go`)
Background service that:
- Polls the outbox table every 5 seconds
- Publishes unpublished events to the message broker
- Marks events as published

### 3. **Event Consumer** (`cmd/event-consumer/main.go`)
Background service that:
- Listens for `tenant.created` events
- Triggers tenant provisioning
- Updates tenant status to `ACTIVE` or `FAILED`

### 4. **Tenant Provisioning Service** (`services/tenant_provisioning.go`)
Handles the actual provisioning logic:
- Updates status to `PROVISIONING`
- Initializes tenant configuration
- Verifies admin user
- Updates status to `ACTIVE`

## Tenant Status Flow

```
PENDING ──► PROVISIONING ──► ACTIVE
                │
                └──────────► FAILED
```

## Running the System

### Prerequisites
1. Database running (MySQL)
2. Configuration loaded (`.env` file)

### Start Services

#### 1. Start Main API Service
```powershell
go run cmd/server/main.go
```

#### 2. Start Outbox Publisher
```powershell
go run cmd/outbox-publisher/main.go
```

#### 3. Start Event Consumer
```powershell
go run cmd/event-consumer/main.go
```

Or use the makefile:
```powershell
make run-publisher
make run-consumer
```

## Development Mode

Currently using **LogBroker** for development, which logs events instead of publishing to a real message broker.

### Testing the Flow

1. **Create a tenant** via API:
```bash
POST /api/v1/admin/tenants
{
  "name": "Test Company",
  "slug": "test-co",
  "plan": "BASIC"
}
```

2. **Check outbox table**:
```sql
SELECT * FROM outbox ORDER BY created_at DESC LIMIT 5;
```

3. **Watch logs**:
- Outbox Publisher logs: Event published
- Event Consumer logs: Tenant provisioning started → completed

4. **Verify tenant status**:
```sql
SELECT id, name, status, created_at FROM tenants ORDER BY created_at DESC LIMIT 5;
```

Status should change: `PENDING` → `PROVISIONING` → `ACTIVE`

## Production Configuration

### Using Kafka

Replace `messaging.NewLogBroker()` with actual Kafka implementation:

```go
// In cmd/outbox-publisher/main.go
broker := messaging.NewKafkaBroker(cfg.Kafka)

// In cmd/event-consumer/main.go
messageConsumer := messaging.NewKafkaConsumer(cfg.Kafka, logger)
```

### Environment Variables

```env
# Message Broker (for future Kafka/RabbitMQ integration)
BROKER_TYPE=log           # log, kafka, rabbitmq
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=tenant-provisioning-consumer

# Outbox Publisher
OUTBOX_BATCH_SIZE=100
OUTBOX_POLL_INTERVAL=5s

# Logging
LOG_LEVEL=info
```

## Monitoring

### Key Metrics to Monitor

1. **Outbox backlog**: Count of unpublished events
```sql
SELECT COUNT(*) FROM outbox WHERE published_at IS NULL;
```

2. **Provisioning failures**: Count of failed tenants
```sql
SELECT COUNT(*) FROM tenants WHERE status = 'FAILED';
```

3. **Average provisioning time**
```sql
SELECT AVG(TIMESTAMPDIFF(SECOND, created_at, updated_at)) 
FROM tenants 
WHERE status = 'ACTIVE';
```

## Troubleshooting

### Tenant Stuck in PENDING

**Cause**: Outbox publisher not running

**Solution**:
1. Check if publisher is running
2. Check outbox table for unpublished events
3. Start the publisher: `go run cmd/outbox-publisher/main.go`

### Tenant Stuck in PROVISIONING

**Cause**: Event consumer not running or provisioning failed

**Solution**:
1. Check event consumer logs for errors
2. Check tenant `failed_reason` field
3. Retry provisioning manually:
```sql
UPDATE tenants SET status = 'PENDING' WHERE id = '<tenant-id>';
```

### Events Not Being Consumed

**Cause**: Event consumer not subscribed or handler not registered

**Solution**:
1. Check consumer logs for subscription confirmation
2. Verify handler is registered for `tenant.created` event
3. Check message broker connectivity

## Manual Operations

### Manually Activate a Stuck Tenant
```sql
UPDATE tenants 
SET status = 'ACTIVE', updated_at = NOW() 
WHERE id = '<tenant-id>' AND status = 'PENDING';
```

### Replay Failed Provisioning
```sql
UPDATE tenants 
SET status = 'PENDING', failed_reason = NULL 
WHERE id = '<tenant-id>' AND status = 'FAILED';
```

### View Outbox Events for a Tenant
```sql
SELECT * FROM outbox 
WHERE aggregate_type = 'tenant' 
  AND aggregate_id = '<tenant-id>'
ORDER BY created_at DESC;
```

## Future Enhancements

1. **Retry Logic**: Implement exponential backoff for failed provisions
2. **Dead Letter Queue**: Move permanently failed events to DLQ
3. **Idempotency**: Add idempotency keys to prevent duplicate provisioning
4. **Observability**: Add OpenTelemetry tracing
5. **Event Sourcing**: Store full event history for audit trail
6. **Real Message Broker**: Integrate Kafka/RabbitMQ/NATS

## Testing

### Unit Tests
```powershell
go test ./messaging/...
go test ./services/...
go test ./consumers/...
```

### Integration Tests
```powershell
go test -tags=integration ./...
```

## References

- [Outbox Pattern](https://microservices.io/patterns/data/transactional-outbox.html)
- [Event-Driven Architecture](https://martinfowler.com/articles/201701-event-driven.html)
- [GORM Documentation](https://gorm.io/docs/)
