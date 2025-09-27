# Mock API responses for rtr-user-auth-service

This folder contains JSON request/response examples you can use to mock the backend for the UI.

Structure:
- docs/mocks/ - individual JSON payloads grouped by endpoint
- This README lists the files and their purpose.

Notes:
- Tenant-scoped routes return errors as `{ "error": "..." }`.
- Admin control-plane routes return structured errors `{ "code": "...", "message": "..." }` (and optional `suggestions`).
- Protected routes require `Authorization: Bearer <token>` and `X-Tenant-ID` headers (the latter is supplied by login).
- Admin routes require a `SUPERADMIN` token and `Idempotency-Key` header for create operations.
- Async tenant creation returns `202 Accepted` for new requests and `200 OK` for cached responses.

Files

Public
- docs/mocks/login.request.json
- docs/mocks/login.response.200.json
- docs/mocks/tenant.settings.get.response.200.json

Protected (tenant-scoped)
- docs/mocks/me.get.response.200.json
- docs/mocks/me.change_password.request.json
- docs/mocks/users.list.response.200.json
- docs/mocks/users.create.request.json
- docs/mocks/users.create.response.201.json
- docs/mocks/tenant.settings.put.request.json
- docs/mocks/tenant.settings.put.response.200.json

Admin (SUPERADMIN)
- docs/mocks/admin/tenants.onboard.request.json
- docs/mocks/admin/tenants.onboard.response.202.json
- docs/mocks/admin/tenants.create.response.200.cached.json
- docs/mocks/admin/tenants.create.no_domain.request.json
- docs/mocks/admin/tenants.create.no_domain.response.202.json
- docs/mocks/admin/tenants.create.no_plan.request.json
- docs/mocks/admin/tenants.create.no_plan.response.202.json
- docs/mocks/admin/tenants.create.error.403.superadmin.json
- docs/mocks/admin/tenants.create.error.409.slug.json
- docs/mocks/admin/tenants.create.error.409.domain.json
- docs/mocks/admin/tenants.create.error.409.idempotency.json
- docs/mocks/admin/tenants.create.error.400.validation.json
- docs/mocks/admin/tenants.create.error.422.no_idempotency.json
- docs/mocks/admin/tenants.get.response.200.json
- docs/mocks/admin/tenants.get.active.response.200.json
- docs/mocks/admin/tenants.get.failed.response.200.json
- docs/mocks/admin/tenants.get.error.404.notfound.json
- docs/mocks/admin/tenants.status.response.200.json
- docs/mocks/admin/tenants.status.active.response.200.json
- docs/mocks/admin/tenants.status.failed.response.200.json

Admin (SUPERADMIN) - New Async Tenant Creation API overview lives in `docs/mocks/admin/README.md` and `docs/mocks/admin/example-usage.md`.

## Curl Examples

All examples assume the service runs locally on `http://localhost:8082`. Replace placeholder tokens, IDs, and request bodies as needed. JSON payloads referenced below are located in the `docs/mocks/` folder.

### Public Endpoints

#### Login (`POST /login`)
```bash
curl -X POST http://localhost:8082/login \
  -H "Content-Type: application/json" \
  --data @docs/mocks/login.request.json
```

### Tenant-Scoped (Authenticated) Endpoints
Use the Bearer token and `X-Tenant-ID` value returned by the login response.

#### Get Current User (`GET /me`)
```bash
curl -X GET http://localhost:8082/me \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: <tenant-id>"
```

#### Change Password (`POST /me/change-password`)
```bash
curl -X POST http://localhost:8082/me/change-password \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: <tenant-id>" \
  -H "Content-Type: application/json" \
  --data @docs/mocks/me.change_password.request.json
```

#### Logout (`POST /logout`)
```bash
curl -X POST http://localhost:8082/logout \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: <tenant-id>"
```

#### List Users (`GET /users`)
```bash
curl -X GET http://localhost:8082/users \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: <tenant-id>"
```

#### Create User (`POST /users`)
```bash
curl -X POST http://localhost:8082/users \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: <tenant-id>" \
  -H "Content-Type: application/json" \
  --data @docs/mocks/users.create.request.json
```

#### Get Tenant Settings (`GET /tenant/settings`)
```bash
curl -X GET http://localhost:8082/tenant/settings \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: <tenant-id>"
```

#### Update Tenant Settings (`PUT /tenant/settings`)
```bash
curl -X PUT http://localhost:8082/tenant/settings \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: <tenant-id>" \
  -H "Content-Type: application/json" \
  --data @docs/mocks/tenant.settings.put.request.json
```

### Admin Control-Plane Endpoints
These endpoints require a SUPERADMIN token and omit tenant context. Always include an `Idempotency-Key` for create operations.
#### Admin Login (`POST /admin/login`)
```bash
curl -X POST http://localhost:8082/admin/login \\
  -H "Content-Type: application/json" \\
  --data @docs/mocks/login.request.json
```
Response includes `PlatformBranding` when the authenticated user has role `SUPERADMIN`.


#### Create Tenant (`POST /tenant/create`)
```bash
curl -X POST http://localhost:8082/tenant/create \
  -H "Authorization: Bearer <superadmin-token>" \
  -H "Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  --data @docs/mocks/admin/tenants.onboard.request.json
```

#### Retry Tenant Creation (Cached response example)
```bash
curl -X POST http://localhost:8082/tenant/create \
  -H "Authorization: Bearer <superadmin-token>" \
  -H "Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  --data @docs/mocks/admin/tenants.onboard.request.json
```
(Second invocation with the same body and key returns the cached payload in `docs/mocks/admin/tenants.create.response.200.cached.json`.)

#### Get Tenant (`GET /tenant/:id`)
```bash
curl -X GET http://localhost:8082/tenant/tnt-0000-1111-2222-333344445555 \
  -H "Authorization: Bearer <superadmin-token>"
```

#### Get Tenant Status (`GET /tenant/:id/status`)
```bash
curl -X GET http://localhost:8082/tenant/tnt-0000-1111-2222-333344445555/status \
  -H "Authorization: Bearer <superadmin-token>"
```

#### Retry Provisioning (`POST /tenant/:id/retry`)
```bash
curl -X POST http://localhost:8082/tenant/tnt-0000-1111-2222-333344445555/retry \
  -H "Authorization: Bearer <superadmin-token>"
```

#### Logout (`POST /admin/logout`)
```bash
curl -X POST http://localhost:8082/admin/logout \
  -H "Authorization: Bearer <superadmin-token>"
```

Refer to the files listed above for sample success and error payloads corresponding to each request.


