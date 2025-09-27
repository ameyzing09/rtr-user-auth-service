# Async Tenant Creation API Examples

This document provides practical examples of how to use the new async tenant creation API.

## Basic Tenant Creation

### Request
```bash
curl -X POST http://localhost:8082/tenant/create \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corporation",
    "domain": "acme.com",
    "admin_name": "Alice Admin",
    "admin_email": "admin@acme.com",
    "plan": "STARTER"
  }'
```

### Response (202 Accepted)
```json
{
  "tenant": {
    "id": "tnt-0000-1111-2222-333344445555",
    "name": "Acme Corporation",
    "domain": "acme.com",
    "slug": "acme-corporation"
  },
  "admin_user_id": "u-admin-1111-2222-3333-aaaaaa",
  "temp_password": "Start1234!",
  "status": "PENDING"
}
```

## Idempotency Example

### Same Request (Same Idempotency-Key)
```bash
curl -X POST http://localhost:8082/tenant/create \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corporation",
    "domain": "acme.com",
    "admin_name": "Alice Admin",
    "admin_email": "admin@acme.com",
    "plan": "STARTER"
  }'
```

### Response (200 OK - Cached)
```json
{
  "tenant": {
    "id": "tnt-0000-1111-2222-333344445555",
    "name": "Acme Corporation",
    "domain": "acme.com",
    "slug": "acme-corporation"
  },
  "admin_user_id": "u-admin-1111-2222-3333-aaaaaa",
  "temp_password": "Start1234!",
  "status": "PENDING"
}
```

## Check Tenant Status

### Request
```bash
curl -X GET http://localhost:8082/tenant/tnt-0000-1111-2222-333344445555/status \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Response
```json
{
  "status": "PENDING"
}
```

## Get Tenant Details

### Request
```bash
curl -X GET http://localhost:8082/tenant/tnt-0000-1111-2222-333344445555 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Response
```json
{
  "id": "tnt-0000-1111-2222-333344445555",
  "name": "Acme Corporation",
  "domain": "acme.com",
  "slug": "acme-corporation",
  "plan": "STARTER",
  "status": "PENDING",
  "created_by": "u-superadmin-0000-1111-2222-333344445555",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

## Error Examples

### Slug Conflict (409)
```json
{
  "code": "TENANT_SLUG_TAKEN",
  "message": "tenant slug already taken",
  "suggestions": [
    "acme-corporation-hq",
    "acme-corporation-io",
    "acme-corporation-team"
  ]
}
```

### Domain Conflict (409)
```json
{
  "code": "DOMAIN_IN_USE",
  "message": "domain already in use"
}
```

### Missing Idempotency Key (422)
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Idempotency-Key header is required"
}
```

### SUPERADMIN Required (403)
```json
{
  "code": "SUPERADMIN_REQUIRED",
  "message": "superadmin required"
}
```

### JSON Validation Error (400)
```json
{
  "error": "name is required, admin_email must be a valid email, plan must be one of [BASIC, STARTER, GROWTH, ENTERPRISE, ON_PREM]"
}
```

## Optional Fields

### Without Domain
```json
{
  "name": "Tech Startup Inc",
  "admin_name": "John Founder",
  "admin_email": "john@techstartup.com",
  "plan": "BASIC"
}
```

### Without Plan (defaults to STARTER)
```json
{
  "name": "Default Plan Corp",
  "domain": "default.com",
  "admin_name": "Jane Admin",
  "admin_email": "jane@default.com"
}
```

## Status Progression

1. **PENDING** - Tenant created, admin user provisioned
2. **PROVISIONING** - Infrastructure being set up
3. **AWAITING_BRANDING** - Waiting for branding configuration
4. **ACTIVE** - Tenant fully provisioned and ready

If provisioning fails, status becomes **FAILED** with a non-empty `failed_reason` field via the GET endpoint.
