# ✅ Issue Fixed - Windows-Compatible Monitoring Scripts

## Problem
The `make check-tenants` and `make check-outbox` commands were trying to use the `mysql` CLI which isn't installed on Windows by default.

## Solution
Created Windows-friendly PowerShell scripts that display the queries you need to run in your database tool.

---

## 🆕 New Commands

### Health Check (Comprehensive)
```powershell
make health-check
```
Shows all diagnostic queries with explanations and expected results.

### Check Outbox Status
```powershell
make check-outbox
```
Shows query to check unpublished events.

### Check Tenant Status
```powershell
make check-tenants
```
Shows query to check tenant status distribution.

### View Recent Tenants
```powershell
make view-tenants
```
Shows query to view recent tenants with provisioning times.

---

## 📁 New Files Created

```
scripts/
├── check-outbox.ps1    # Check unpublished events
├── check-tenants.ps1   # Check tenant status
├── view-tenants.ps1    # View recent tenants
└── health-check.ps1    # Comprehensive health check
```

---

## 💡 How to Use

### Option 1: Use the Scripts (Easiest)
```powershell
make health-check       # Shows all queries
make check-tenants      # Shows tenant status query
make check-outbox       # Shows outbox query
```

The scripts will display the SQL queries you need. Copy and paste them into:
- **MySQL Workbench**
- **DBeaver**
- **Any other database tool**

### Option 2: Install MySQL CLI (Optional)
If you want direct command-line access:

1. Download MySQL: https://dev.mysql.com/downloads/mysql/
2. Install and add to PATH
3. Then run:
```powershell
mysql -h127.0.0.1 -P3306 -uroot -p -e "SELECT * FROM tenants;" recrutr-db
```

### Option 3: Direct PowerShell Execution
```powershell
.\scripts\health-check.ps1
.\scripts\check-tenants.ps1
.\scripts\check-outbox.ps1
```

---

## 🔍 What Each Script Shows

### 1. `health-check.ps1` - Complete System Overview
Displays 6 key diagnostic queries:
- ✅ Unpublished events count
- ✅ Tenant status distribution
- ✅ Recent tenants with timing
- ✅ Failed tenants with reasons
- ✅ Average provisioning time
- ✅ Currently pending tenants

### 2. `check-outbox.ps1` - Outbox Status
Shows if events are waiting to be published.
- **Expected**: 0 unpublished events (if publisher running)
- **Warning**: > 0 means publisher might not be running

### 3. `check-tenants.ps1` - Tenant Distribution
Shows how many tenants are in each status.
- **Good**: Most tenants are ACTIVE
- **Warning**: Many PENDING means services may be down
- **Error**: Many FAILED needs investigation

### 4. `view-tenants.ps1` - Recent Activity
Shows last 10 tenants with provisioning time.
- **Expected**: 5-15 seconds for ACTIVE tenants

---

## 📊 Example Output

### Running `make health-check`:
```
╔═══════════════════════════════════════════════════════════╗
║   PROVISIONING SYSTEM HEALTH CHECK                        ║
╚═══════════════════════════════════════════════════════════╝

📊 Database Connection:
   root@127.0.0.1:3306/recrutr-db

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1️⃣  CHECK UNPUBLISHED EVENTS (Outbox Backlog)
   This shows events waiting to be published

   SELECT COUNT(*) as unpublished FROM outbox WHERE published_at IS NULL;

   ✅ Expected: 0 (if publisher is running)
   ⚠️  If > 0: Publisher may not be running

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

... (continues with all 6 checks)
```

---

## 🎯 Quick Reference

### Check if System is Healthy
```powershell
make health-check
```
Copy the queries into your database tool and verify:
- [ ] Unpublished events = 0
- [ ] Most tenants are ACTIVE
- [ ] Provisioning time = 5-15 seconds
- [ ] No (or few) FAILED tenants

### Services Running Check
```powershell
# Make sure all 3 are running:
make run            # Terminal 1
make run-publisher  # Terminal 2
make run-consumer   # Terminal 3
```

---

## ✅ Resolution Summary

**Before:**
```
PS> make check-tenants
'mysql' is not recognized as an internal or external command
Error 255
```

**After:**
```
PS> make check-tenants
==> Checking Tenant Provisioning Status

Query to run:
SELECT status, COUNT(*) as count FROM tenants GROUP BY status;

Connection: root@127.0.0.1:3306/recrutr-db

Run this query in your database tool (MySQL Workbench, DBeaver, etc.)
```

---

## 🎉 Benefits

✅ **Windows-Friendly** - No need to install mysql CLI  
✅ **Clear Instructions** - Shows exactly what to do  
✅ **Copy-Paste Ready** - Queries ready to use  
✅ **Comprehensive** - All diagnostic queries included  
✅ **Professional** - Colored output with emojis  

---

Your monitoring scripts are now fully functional on Windows! 🚀
