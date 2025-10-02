# Provisioning System Health Check
# Usage: .\scripts\health-check.ps1

# Load environment variables from .env if it exists
if (Test-Path .env) {
    Get-Content .env | ForEach-Object {
        if ($_ -match '^([^#][^=]+)=(.*)$') {
            [Environment]::SetEnvironmentVariable($matches[1], $matches[2], 'Process')
        }
    }
}

$DB_HOST = if ($env:DB_HOST) { $env:DB_HOST } else { "127.0.0.1" }
$DB_PORT = if ($env:DB_PORT) { $env:DB_PORT } else { "3306" }
$DB_USER = if ($env:DB_USER) { $env:DB_USER } else { "root" }
$DB_PASSWORD = if ($env:DB_PASSWORD) { $env:DB_PASSWORD } else { "secret" }
$DB_NAME = if ($env:DB_NAME) { $env:DB_NAME } else { "recrutr" }

Write-Host "╔═══════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║   PROVISIONING SYSTEM HEALTH CHECK                        ║" -ForegroundColor Cyan
Write-Host "╚═══════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

Write-Host "📊 Database Connection:" -ForegroundColor Yellow
Write-Host "   ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}" -ForegroundColor Gray
Write-Host ""

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-Host ""

Write-Host "1️⃣  CHECK UNPUBLISHED EVENTS (Outbox Backlog)" -ForegroundColor Cyan
Write-Host "   This shows events waiting to be published" -ForegroundColor Gray
Write-Host ""
Write-Host "   SELECT COUNT(*) as unpublished FROM outbox WHERE published_at IS NULL;" -ForegroundColor White
Write-Host ""
Write-Host "   ✅ Expected: 0 (if publisher is running)" -ForegroundColor Green
Write-Host "   ⚠️  If > 0: Publisher may not be running" -ForegroundColor Yellow
Write-Host ""

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-Host ""

Write-Host "2️⃣  TENANT STATUS DISTRIBUTION" -ForegroundColor Cyan
Write-Host "   Shows how many tenants are in each status" -ForegroundColor Gray
Write-Host ""
Write-Host "   SELECT status, COUNT(*) as count FROM tenants GROUP BY status;" -ForegroundColor White
Write-Host ""
Write-Host "   ✅ Most should be ACTIVE" -ForegroundColor Green
Write-Host "   ⚠️  If many PENDING: Check services" -ForegroundColor Yellow
Write-Host "   ❌ If many FAILED: Check logs" -ForegroundColor Red
Write-Host ""

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-Host ""

Write-Host "3️⃣  RECENT TENANTS (Last 10)" -ForegroundColor Cyan
Write-Host "   Shows recent tenants with provisioning time" -ForegroundColor Gray
Write-Host ""
Write-Host "   SELECT id, name, status," -ForegroundColor White
Write-Host "          TIMESTAMPDIFF(SECOND, created_at, updated_at) as provision_seconds" -ForegroundColor White
Write-Host "   FROM tenants" -ForegroundColor White
Write-Host "   ORDER BY created_at DESC LIMIT 10;" -ForegroundColor White
Write-Host ""
Write-Host "   ✅ Expected: 5-15 seconds for ACTIVE tenants" -ForegroundColor Green
Write-Host ""

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-Host ""

Write-Host "4️⃣  FAILED TENANTS (With Reasons)" -ForegroundColor Cyan
Write-Host "   Shows tenants that failed provisioning" -ForegroundColor Gray
Write-Host ""
Write-Host "   SELECT id, name, status, failed_reason, created_at" -ForegroundColor White
Write-Host "   FROM tenants" -ForegroundColor White
Write-Host "   WHERE status = 'FAILED'" -ForegroundColor White
Write-Host "   ORDER BY created_at DESC;" -ForegroundColor White
Write-Host ""
Write-Host "   ⚠️  Check failed_reason for details" -ForegroundColor Yellow
Write-Host ""

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-Host ""

Write-Host "5️⃣  AVERAGE PROVISIONING TIME" -ForegroundColor Cyan
Write-Host "   Shows average time to provision" -ForegroundColor Gray
Write-Host ""
Write-Host "   SELECT" -ForegroundColor White
Write-Host "       AVG(TIMESTAMPDIFF(SECOND, created_at, updated_at)) as avg_seconds," -ForegroundColor White
Write-Host "       MIN(TIMESTAMPDIFF(SECOND, created_at, updated_at)) as min_seconds," -ForegroundColor White
Write-Host "       MAX(TIMESTAMPDIFF(SECOND, created_at, updated_at)) as max_seconds" -ForegroundColor White
Write-Host "   FROM tenants" -ForegroundColor White
Write-Host "   WHERE status = 'ACTIVE';" -ForegroundColor White
Write-Host ""
Write-Host "   ✅ Expected avg: 5-15 seconds" -ForegroundColor Green
Write-Host ""

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-Host ""

Write-Host "6️⃣  PENDING TENANTS (Waiting for Provisioning)" -ForegroundColor Cyan
Write-Host "   Shows tenants that are waiting" -ForegroundColor Gray
Write-Host ""
Write-Host "   SELECT id, name, status, created_at," -ForegroundColor White
Write-Host "          TIMESTAMPDIFF(SECOND, created_at, NOW()) as waiting_seconds" -ForegroundColor White
Write-Host "   FROM tenants" -ForegroundColor White
Write-Host "   WHERE status IN ('PENDING', 'PROVISIONING')" -ForegroundColor White
Write-Host "   ORDER BY created_at DESC;" -ForegroundColor White
Write-Host ""
Write-Host "   ⚠️  If waiting > 60s: Check services are running" -ForegroundColor Yellow
Write-Host ""

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-Host ""

Write-Host "📋 HOW TO RUN THESE QUERIES:" -ForegroundColor Yellow
Write-Host ""
Write-Host "   Option 1: Copy queries into MySQL Workbench or DBeaver" -ForegroundColor Green
Write-Host "   Option 2: Install mysql CLI and run from terminal" -ForegroundColor Green
Write-Host "   Option 3: Use scripts in ./scripts/ folder" -ForegroundColor Green
Write-Host ""

Write-Host "🔧 QUICK COMMANDS:" -ForegroundColor Yellow
Write-Host ""
Write-Host "   make check-outbox       # Check unpublished events" -ForegroundColor White
Write-Host "   make check-tenants      # Check tenant status" -ForegroundColor White
Write-Host "   make view-tenants       # View recent tenants" -ForegroundColor White
Write-Host ""

Write-Host "🚀 SERVICES STATUS:" -ForegroundColor Yellow
Write-Host ""
Write-Host "   Check if these are running:" -ForegroundColor Gray
Write-Host "   • Main API Server (make run)" -ForegroundColor White
Write-Host "   • Outbox Publisher (make run-publisher)" -ForegroundColor White
Write-Host "   • Event Consumer (make run-consumer)" -ForegroundColor White
Write-Host ""

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-Host ""
Write-Host "✅ Health check queries ready to run!" -ForegroundColor Green
Write-Host ""
