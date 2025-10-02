# ✅ Solution: Retry Provisioning for Stuck Tenants

## Your Question
> "Now one tenant is showing pending before this was implemented how can we push it in queue or any solution?"

## Answer: Yes! Here's How to Fix It

You can manually create an outbox event that will trigger provisioning for any tenant stuck in PENDING status.

---

## 🚀 Quick Solution (3 Steps)

### Step 1: Run the Retry Script

```powershell
# For all pending tenants:
.\scripts\retry-provisioning.ps1 -AllPending

# Or for a specific tenant:
.\scripts\retry-provisioning.ps1 -TenantID "your-tenant-id-here"
```

### Step 2: Copy the SQL Query Shown

The script will display a query. Copy it.

### Step 3: Run in Your Database Tool

Paste and run the query in MySQL Workbench or DBeaver.

**Done!** The tenant will be provisioned within 5-15 seconds.

---

## 📋 Example: Fix a Specific Tenant

### 1. Find your tenant ID:
```sql
SELECT id, name, status FROM tenants WHERE status = 'PENDING';
```

Let's say you get: `abc-123-def-456`

### 2. Run the script:
```powershell
.\scripts\retry-provisioning.ps1 -TenantID "abc-123-def-456"
```

### 3. The script shows you this query:
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
WHERE id = 'abc-123-def-456';
```

###4. Copy and run it in MySQL Workbench

### 5. Wait 10 seconds and check:
```sql
SELECT id, name, status FROM tenants WHERE id = 'abc-123-def-456';
```

**Result:** Status should now be `ACTIVE`! ✅

---

## 📋 Example: Fix ALL Pending Tenants

### 1. Run the script:
```powershell
.\scripts\retry-provisioning.ps1 -AllPending
```

### 2. Run the query shown:
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
  AND id NOT IN (SELECT aggregate_id FROM outbox WHERE aggregate_type = 'tenant');
```

This will:
- ✅ Find ALL tenants with PENDING status
- ✅ Skip tenants that already have outbox events
- ✅ Create events for the rest
- ✅ Provision all of them automatically

---

## ⚠️ Prerequisites

**IMPORTANT:** These services MUST be running:

```powershell
# Terminal 1 - API Server
make run

# Terminal 2 - Outbox Publisher (REQUIRED!)
make run-publisher

# Terminal 3 - Event Consumer (REQUIRED!)
make run-consumer
```

Without the publisher and consumer, the events won't be processed!

---

## 🔍 What Happens Behind the Scenes

```
1. You run INSERT query
   ↓
2. Event created in outbox table
   ↓ (0-5 seconds)
3. Publisher reads and publishes event
   ↓ (1-2 seconds)
4. Consumer receives event
   ↓ (1-2 seconds)
5. Provisioning service provisions tenant
   ↓ (1-2 seconds)
6. Tenant status updated to ACTIVE ✅

Total: 5-15 seconds
```

---

## 📊 Verify It Worked

### Check the tenant status:
```sql
SELECT id, name, status, updated_at
FROM tenants 
WHERE id = 'your-tenant-id'
ORDER BY updated_at DESC;
```

### Check the outbox event:
```sql
SELECT * FROM outbox 
WHERE aggregate_id = 'your-tenant-id'
ORDER BY created_at DESC LIMIT 1;
```

Should see:
- ✅ `published_at` has a timestamp
- ✅ Tenant status is `ACTIVE`
- ✅ `updated_at` is recent

---

## 🛠️ Troubleshooting

### Problem: Tenant still PENDING after 30 seconds

**Check 1:** Are services running?
```powershell
Get-Process | Where-Object { $_.ProcessName -like "*outbox*" -or $_.ProcessName -like "*consumer*" }
```

**Check 2:** Is event published?
```sql
SELECT * FROM outbox WHERE published_at IS NULL;
```

**Fix:** Start the missing services
```powershell
make run-publisher  # Terminal 2
make run-consumer   # Terminal 3
```

### Problem: Tenant changed to FAILED

**Check reason:**
```sql
SELECT id, name, failed_reason 
FROM tenants 
WHERE status = 'FAILED';
```

**Fix and retry:**
```sql
-- Reset to PENDING
UPDATE tenants SET status = 'PENDING', failed_reason = NULL WHERE id = 'tenant-id';

-- Then create new outbox event (use script again)
```

---

## 📁 Files Created

- ✅ `scripts/retry-provisioning.ps1` - Interactive script
- ✅ `docs/RETRY_PROVISIONING.md` - Detailed guide
- ✅ `makefile` - Added retry commands

---

## 🎯 Commands Summary

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

# View recent tenants
make view-tenants
```

---

## ✅ Success Checklist

After running the retry:

- [ ] Event appears in outbox table
- [ ] Event has `published_at` timestamp within 5 seconds
- [ ] Tenant status changes to PROVISIONING
- [ ] Tenant status changes to ACTIVE
- [ ] `updated_at` timestamp is recent
- [ ] No errors in publisher/consumer logs

---

## 💡 Pro Tips

1. **Always verify services are running first**
2. **Monitor the logs** while retrying to see real-time progress
3. **Be patient** - wait at least 15 seconds before checking again
4. **Batch process** if you have many stuck tenants

---

## 📚 More Information

- Full guide: `docs/RETRY_PROVISIONING.md`
- System overview: `docs/PROVISIONING_SYSTEM.md`
- Quick start: `docs/QUICK_START_PROVISIONING.md`

---

## 🎉 Summary

**Your stuck tenant can be fixed in 3 simple steps:**

1. Run: `.\scripts\retry-provisioning.ps1 -TenantID "your-id"`
2. Copy the SQL query shown
3. Run it in your database tool

**The tenant will be provisioned automatically within 5-15 seconds!** ✅

---

**Problem Solved!** 🚀
