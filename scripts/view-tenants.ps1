# View Recent Tenants
# Usage: .\scripts\view-tenants.ps1

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

Write-Host "==> Recent Tenants (Last 10)" -ForegroundColor Cyan
Write-Host ""

$query = @"
SELECT 
    id, 
    name, 
    status, 
    TIMESTAMPDIFF(SECOND, created_at, updated_at) as provision_seconds,
    created_at,
    updated_at
FROM tenants 
ORDER BY created_at DESC 
LIMIT 10;
"@

Write-Host "Query to run:" -ForegroundColor Yellow
Write-Host $query -ForegroundColor Gray
Write-Host ""
Write-Host "Connection: ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}" -ForegroundColor Gray
Write-Host ""
Write-Host "Copy and paste this query into your database tool" -ForegroundColor Green
