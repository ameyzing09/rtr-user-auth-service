# Admin Tenant Management API Mocks

This directory contains mock responses for the admin tenant management API endpoints.

## Files Overview

### Request Examples
- `tenants.onboard.request.json` - Basic tenant creation request
- `tenants.create.no_domain.request.json` - Request without domain (optional)
- `tenants.create.no_plan.request.json` - Request without plan (defaults to STARTER)

### Success Responses
- `tenants.onboard.response.202.json` - Successful tenant creation (202 Accepted)
- `tenants.create.response.200.cached.json` - Cached create response (200 OK)
- `tenants.create.no_domain.response.202.json` - Response for no-domain request
- `tenants.create.no_plan.response.202.json` - Response for no-plan request
- `tenants.get.response.200.json` - Basic tenant details
- `tenants.get.active.response.200.json` - Active tenant details
- `tenants.get.failed.response.200.json` - Failed tenant details (with failed_reason)
- `tenants.status.response.200.json` - Pending status (no steps yet)
- `tenants.status.active.response.200.json` - Active status with example steps
- `tenants.status.failed.response.200.json` - Failed status with example steps

### Error Responses
- `tenants.create.error.403.superadmin.json` - SUPERADMIN role required
- `tenants.create.error.409.slug.json` - Slug conflict with suggestions
- `tenants.create.error.409.domain.json` - Domain already in use
- `tenants.create.error.409.idempotency.json` - Idempotency key reuse with different payload
- `tenants.create.error.400.validation.json` - JSON/body validation errors
- `tenants.create.error.422.no_idempotency.json` - Missing idempotency key header
- `tenants.get.error.404.notfound.json` - Tenant not found

### Documentation
- `example-usage.md` - Comprehensive usage examples with curl commands
- `README.md` - This file

## API Endpoints

### POST /tenant/create
Creates a new tenant with async provisioning.

**Headers Required:**
- `Authorization: Bearer <jwt_token>` (SUPERADMIN role)
- `Idempotency-Key: <uuid_or_token>`
- `Content-Type: application/json`

**Request Body:**
```json
{
  "name": "Company Name",
  "domain": "company.com",
  "admin_name": "Admin Name",
  "admin_email": "admin@company.com",
  "plan": "STARTER"
}
```

**Response Codes:**
- `202 Accepted` - New tenant creation started
- `200 OK` - Cached response (same Idempotency-Key)
- `409 Conflict` - Slug/domain conflict or idempotency key reuse
- `400 Bad Request` - JSON/body validation errors
- `422 Unprocessable Entity` - Missing Idempotency-Key header
- `403 Forbidden` - SUPERADMIN role required

### GET /tenant/:id
Retrieves tenant details.

**Headers Required:**
- `Authorization: Bearer <jwt_token>` (SUPERADMIN role)

**Response:** Tenant details including status, plan, timestamps, and optional `created_by` / `failed_reason`.

### GET /tenant/:id/status
Retrieves tenant provisioning status.

**Headers Required:**
- `Authorization: Bearer <jwt_token>` (SUPERADMIN role)

**Response:** Status and optional provisioning steps.

### POST /tenant/:id/retry
Retries failed tenant provisioning.

**Headers Required:**
- `Authorization: Bearer <jwt_token>` (SUPERADMIN role)

**Response:** `202 Accepted` - Retry request accepted and outbox event enqueued.

## Tenant Status Values

- `PENDING` - Tenant created, awaiting provisioning
- `PROVISIONING` - Infrastructure being set up
- `AWAITING_BRANDING` - Waiting for branding configuration
- `ACTIVE` - Tenant fully provisioned and ready
- `FAILED` - Provisioning failed (check failed_reason)
- `SUSPENDED` - Tenant suspended
- `DELETED` - Tenant marked for deletion

## Plan Values

- `BASIC` - Basic plan
- `STARTER` - Starter plan (default)
- `GROWTH` - Growth plan
- `ENTERPRISE` - Enterprise plan
- `ON_PREM` - On-premises plan

## Error Codes

- `SUPERADMIN_REQUIRED` - SUPERADMIN role required
- `TENANT_SLUG_TAKEN` - Slug already taken (includes suggestions)
- `DOMAIN_IN_USE` - Domain already in use
- `IDEMPOTENCY_KEY_REUSE_DIFFERENT_REQUEST` - Idempotency key reuse with different payload
- `VALIDATION_ERROR` - Input validation failed (missing idempotency key)
- `TENANT_NOT_FOUND` - Tenant not found
- `INTERNAL_ERROR` - Server error

## Usage Tips

1. **Idempotency**: Always include a unique `Idempotency-Key` header for create operations.
2. **Error Handling**: Check the `code` field in admin error responses for specific error types.
3. **Status Polling**: Use the status endpoint to check provisioning progress.
4. **Retry Logic**: Use the retry endpoint for failed provisioning attempts.
5. **Slug Suggestions**: When slug conflicts occur, use the provided suggestions.
