# Retry Provisioning for Existing/Stuck Tenants
param(
    [string]$TenantID = "",
    [switch]$AllPending = $false
)

# Load environment
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

Write-Host "==============================================================" -ForegroundColor Cyan
Write-Host "   RETRY TENANT PROVISIONING" -ForegroundColor Cyan
Write-Host "==============================================================" -ForegroundColor Cyan
Write-Host ""

if ($TenantID -eq "" -and -not $AllPending) {
    Write-Host "ERROR: Please provide a TenantID or use -AllPending flag" -ForegroundColor Red
    Write-Host ""
    Write-Host "Usage Examples:" -ForegroundColor Yellow
    Write-Host "  .\scripts\retry-provisioning.ps1 -TenantID 'abc-123-def'" -ForegroundColor White
    Write-Host "  .\scripts\retry-provisioning.ps1 -AllPending" -ForegroundColor White
    Write-Host ""
    exit 1
}

Write-Host "Database:" $DB_USER"@"$DB_HOST":"$DB_PORT"/"$DB_NAME -ForegroundColor Gray
Write-Host ""
Write-Host "==============================================================" -ForegroundColor DarkGray
Write-Host ""

if ($AllPending) {
    Write-Host "RETRY ALL PENDING TENANTS" -ForegroundColor Cyan
    Write-Host ""
    
    Write-Host "Step 1: Find all pending tenants" -ForegroundColor Green
    Write-Host ""
    Write-Host "SELECT id, name, created_at FROM tenants WHERE status = 'PENDING';" -ForegroundColor White
    Write-Host ""
    
    Write-Host "Step 2: Create outbox events for all pending tenants" -ForegroundColor Green
    Write-Host ""
    $query = @"
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
"@
    Write-Host $query -ForegroundColor White
    Write-Host ""
    
} else {
    Write-Host "RETRY PROVISIONING FOR TENANT: $TenantID" -ForegroundColor Cyan
    Write-Host ""
    
    Write-Host "Step 1: Verify tenant exists" -ForegroundColor Green
    Write-Host ""
    Write-Host "SELECT id, name, status, created_at FROM tenants WHERE id = '$TenantID';" -ForegroundColor White
    Write-Host ""
    
    Write-Host "Step 2: Create outbox event to trigger provisioning" -ForegroundColor Green
    Write-Host ""
    $query = @"
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
WHERE id = '$TenantID';
"@
    Write-Host $query -ForegroundColor White
    Write-Host ""
}

Write-Host "==============================================================" -ForegroundColor DarkGray
Write-Host ""

Write-Host "WHAT HAPPENS NEXT:" -ForegroundColor Yellow
Write-Host "  1. Run the query above in your database tool" -ForegroundColor White
Write-Host "  2. Outbox publisher picks it up (within 5 seconds)" -ForegroundColor White
Write-Host "  3. Event consumer provisions the tenant" -ForegroundColor White
Write-Host "  4. Tenant status changes to ACTIVE" -ForegroundColor White
Write-Host ""
Write-Host "Expected time: 5-15 seconds" -ForegroundColor Green
Write-Host ""

Write-Host "==============================================================" -ForegroundColor DarkGray
Write-Host ""

Write-Host "IMPORTANT - Services must be running:" -ForegroundColor Yellow
Write-Host "  - Outbox Publisher: make run-publisher" -ForegroundColor White
Write-Host "  - Event Consumer: make run-consumer" -ForegroundColor White
Write-Host ""

Write-Host "Copy the query above and run it in MySQL Workbench or DBeaver!" -ForegroundColor Green
Write-Host ""
