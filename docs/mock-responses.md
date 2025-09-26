# Mock API responses for rtr-user-auth-service

This folder contains JSON request/response examples you can use to mock the backend for the UI.

Structure:
- docs/mocks/ - individual JSON payloads grouped by endpoint
- This README lists the files and their purpose.

Notes:
- All error responses use the single-key shape: `{ "error": "<message>" }`.
- Protected routes require `Authorization: Bearer <token>` header.
- Login response includes an `X-Tenant-ID` header the server sets; UI should persist the tenant id if needed.

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
- docs/mocks/admin/tenants.onboard.response.201.json
- docs/mocks/admin/tenants.get.response.200.json
- docs/mocks/admin/tenants.by_domain.response.200.json

How to use
- Point your mock server at the appropriate JSON files. Example frameworks: json-server, mockoon, wiremock.

If you want additional variants (pagination for users, 400/401/403 examples as separate files, or sample OpenAPI/Swagger), tell me which ones and I will add them.
