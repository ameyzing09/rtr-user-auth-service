# 🔄 Retry Provisioning Guide

## Problem: Tenant Stuck in PENDING

You have tenants that were created **before** the provisioning system was implemented, or tenants that failed to provision.

## Solutions

### Solution 1: Retry Specific Tenant (Recommended)

#### Using PowerShell Script:
```powershell
# Replace with your actual tenant ID
.\scripts\retry-provisioning.ps1 -TenantID "abc-123-def-456"
```

#### Using Make Command:
```powershell
make retry-tenant TENANT_ID=abc-123-def-456
```

The script will show you the SQL query to run. Copy it and execute in your database tool.

---

### Solution 2: Retry All Pending Tenants

#### Using PowerShell Script:
```powershell
.\scripts\retry-provisioning.ps1 -AllPending
```

#### Using Make Command:
```powershell
make retry-all-pending
```

---

## Step-by-Step Manual Process

### 1. Find Your Pending Tenant(s)

```sql
SELECT id, name, status, created_at,
       TIMESTAMPDIFF(SECOND, created_at, NOW()) as waiting_seconds
FROM tenants 
WHERE status = 'PENDING'
ORDER BY created_at DESC;
```

### 2. Create Outbox Event for the Tenant

Replace `YOUR_TENANT_ID` with the actual ID:

```sql
INSERT INTO outbox (aggregate_type, aggregate_id, type, payload, created_at)
SELECT 
    'tenant',
    id,
    'tenant.created',
    JSON_OBJECT(
        'v', 1,
        'tenantId', id,
        'name', name,
        'plan', IFNULL(plan, 'BASIC'),
        'creatorUserId', IFNULL(created_by, ''),
        'createdAt', DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ')
    ),
    NOW()
FROM tenants
WHERE id = 'YOUR_TENANT_ID';
```

### 3. Verify Event Was Created

```sql
SELECT * FROM outbox 
WHERE aggregate_id = 'YOUR_TENANT_ID' 
ORDER BY created_at DESC LIMIT 1;
```

### 4. Wait for Provisioning (5-15 seconds)

The system will automatically:
- ✅ Outbox publisher picks up the event (within 5s)
- ✅ Event consumer provisions the tenant (1-2s)
- ✅ Tenant status updates to ACTIVE

### 5. Verify Tenant is Now Active

```sql
SELECT id, name, status, updated_at
FROM tenants 
WHERE id = 'YOUR_TENANT_ID';
```

---

## Batch Process: Retry All Pending Tenants

If you have multiple pending tenants, use this query:

```sql
INSERT INTO outbox (aggregate_type, aggregate_id, type, payload, created_at)
SELECT 
    'tenant',
    id,
    'tenant.created',
    JSON_OBJECT(
        'v', 1,
        'tenantId', id,
        'name', name,
        'plan', IFNULL(plan, 'BASIC'),
        'creatorUserId', IFNULL(created_by, ''),
        'createdAt', DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ')
    ),
    NOW()
FROM tenants
WHERE status = 'PENDING'
  AND id NOT IN (
      SELECT aggregate_id 
      FROM outbox 
      WHERE aggregate_type = 'tenant'
  );
```

This will:
- ✅ Find all tenants with PENDING status
- ✅ Skip tenants that already have events in outbox
- ✅ Create events for all others

---

## Prerequisites

**Make sure these services are running:**

```powershell
# Terminal 1 - Main API
make run

# Terminal 2 - Outbox Publisher ⚠️ REQUIRED
make run-publisher

# Terminal 3 - Event Consumer ⚠️ REQUIRED
make run-consumer
```

**Without the publisher and consumer running, the events won't be processed!**

---

## Timeline

```
Run INSERT query → Event created in outbox
    ↓ (0-5 seconds)
Publisher picks up event
    ↓ (1-2 seconds)
Consumer provisions tenant
    ↓ (1-2 seconds)
Tenant status = ACTIVE ✅

Total: 5-15 seconds
```

---

## Troubleshooting

### Event Created But Tenant Still PENDING

**Check 1: Is publisher running?**
```powershell
# Look for this process
Get-Process | Where-Object { $_.ProcessName -like "*outbox*" }
```

**Check 2: Check outbox table**
```sql
SELECT * FROM outbox 
WHERE published_at IS NULL
ORDER BY created_at DESC;
```

If `published_at` is NULL, publisher hasn't run yet.

**Solution:** Start the publisher
```powershell
make run-publisher
```

---

### Event Published But Tenant Still PENDING

**Check 1: Is consumer running?**
```powershell
# Look for this process
Get-Process | Where-Object { $_.ProcessName -like "*event-consumer*" }
```

**Check 2: Check consumer logs**
Look for errors in the terminal where consumer is running.

**Solution:** Start the consumer
```powershell
make run-consumer
```

---

### Tenant Status Changed to FAILED

**Check the failure reason:**
```sql
SELECT id, name, status, failed_reason, created_at
FROM tenants 
WHERE status = 'FAILED';
```

**Fix the issue, then retry:**
```sql
-- Reset to PENDING
UPDATE tenants 
SET status = 'PENDING', failed_reason = NULL 
WHERE id = 'YOUR_TENANT_ID';

-- Then create new outbox event (use query from above)
```

---

## Quick Reference

### Commands
```powershell
# Retry one tenant
.\scripts\retry-provisioning.ps1 -TenantID "tenant-id"
make retry-tenant TENANT_ID=tenant-id

# Retry all pending
.\scripts\retry-provisioning.ps1 -AllPending
make retry-all-pending

# Check status
make check-tenants
make health-check
```

### Key Queries
```sql
-- Find pending tenants
SELECT * FROM tenants WHERE status = 'PENDING';

-- Check unpublished events
SELECT * FROM outbox WHERE published_at IS NULL;

-- Check tenant status
SELECT id, name, status FROM tenants WHERE id = 'YOUR_ID';
```

---

## Example: Complete Workflow

```powershell
# 1. Find pending tenants
make check-tenants

# 2. Copy tenant ID from the output

# 3. Retry provisioning for that tenant
.\scripts\retry-provisioning.ps1 -TenantID "abc-123-def"

# 4. Copy the SQL query shown

# 5. Run it in MySQL Workbench/DBeaver

# 6. Wait 10 seconds

# 7. Verify tenant is now ACTIVE
make check-tenants
```

---

## ✅ Success Indicators

After running the retry:

1. ✅ Event appears in outbox table
2. ✅ Event gets `published_at` timestamp within 5 seconds
3. ✅ Tenant status changes to PROVISIONING
4. ✅ Tenant status changes to ACTIVE
5. ✅ `updated_at` timestamp is recent

---

## 🎯 Pro Tips

1. **Always check services are running first**
   ```powershell
   Get-Process | Select-String "outbox|consumer"
   ```

2. **Monitor the logs while retrying**
   - Watch Terminal 2 (publisher) for event publishing
   - Watch Terminal 3 (consumer) for provisioning

3. **Be patient**
   - Wait at least 15 seconds before checking again
   - The system polls every 5 seconds

4. **Batch process if multiple tenants**
   - Use the "retry all pending" query
   - Processes all at once

---

**Your tenant will be provisioned automatically after creating the outbox event!** 🚀
