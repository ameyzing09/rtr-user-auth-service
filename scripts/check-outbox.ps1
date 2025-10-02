# Check Outbox Status
# Usage: .\scripts\check-outbox.ps1

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

Write-Host "==> Checking Outbox Status" -ForegroundColor Cyan
Write-Host ""

$query = "SELECT COUNT(*) as unpublished FROM outbox WHERE published_at IS NULL;"

Write-Host "Query to run:" -ForegroundColor Yellow
Write-Host $query -ForegroundColor Gray
Write-Host ""
Write-Host "Connection: ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}" -ForegroundColor Gray
Write-Host ""
Write-Host "Run this query in your database tool (MySQL Workbench, DBeaver, etc.)" -ForegroundColor Green
Write-Host ""
Write-Host "Or install mysql CLI: https://dev.mysql.com/downloads/mysql/" -ForegroundColor Yellow
Write-Host "Then run: mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASSWORD} -e `"${query}`" ${DB_NAME}" -ForegroundColor Gray
