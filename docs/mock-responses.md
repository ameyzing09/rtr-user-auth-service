# Mock Payload Directory

Use this guide to locate the JSON fixtures that mirror live responses. File paths are relative to `docs/mocks/`.

## Legend

| Group | Description |
| --- | --- |
| Public | Anonymous login flows. |
| Tenant | Authenticated tenant-scoped operations (requires `Authorization` + `X-Tenant-ID`). |
| Admin | SUPERADMIN control-plane commands. |
| Errors | Generic error payloads used across flows. |

## Public

| File | Notes |
| --- | --- |
| `login.request.json` | Sample login body. |
| `login.response.200.json` | Tenant login success payload. |

## Tenant-Scoped

| File | Notes |
| --- | --- |
| `me.get.response.200.json` | `/me` profile response. |
| `me.change_password.request.json` | Change password request body. |
| `users.list.response.200.json` | `/users` listing. |
| `users.create.request.json` | Create user body. |
| `users.create.response.201.json` | Create user success. |
| `tenant.settings.get.response.200.json` | GET tenant settings. |
| `tenant.settings.put.request.json` | PUT settings body. |
| `tenant.settings.put.response.200.json` | Updated settings echo. |

## Control-Plane (SUPERADMIN)

| File | Notes |
| --- | --- |
| `admin/login.response.200.json` | `/admin/login` response with `PlatformBranding`. |
| `admin/tenants.list.response.200.json` | `/admin/tenants` list payload. |
| `admin/tenants.onboard.request.json` | POST `/tenant/create` request. |
| `admin/tenants.onboard.response.202.json` | Async create response (202). |
| `admin/tenants.create.response.200.cached.json` | Replayed response when idempotency hits. |
| `admin/tenants.create.no_domain.request.json` | Create without domain. |
| `admin/tenants.create.no_domain.response.202.json` | Response for above. |
| `admin/tenants.create.no_plan.request.json` | Create without plan. |
| `admin/tenants.create.no_plan.response.202.json` | Response (defaults to STARTER). |
| `admin/tenants.get.response.200.json` | `/tenant/:id` base payload. |
| `admin/tenants.get.active.response.200.json` | Active tenant variant. |
| `admin/tenants.get.failed.response.200.json` | Failed tenant variant. |
| `admin/tenants.status.response.200.json` | `/tenant/:id/status` baseline. |
| `admin/tenants.status.active.response.200.json` | Sample with steps. |
| `admin/tenants.status.failed.response.200.json` | Failed provisioning steps. |
| `admin/tenants.create.error.*.json` | Error variants (403/409/422). |
| `admin/tenants.onboard.response.error.409.json` | Domain conflict. |
| `admin/tenants.get.error.404.notfound.json` | Missing tenant. |

## Errors

| File | HTTP Code |
| --- | --- |
| `errors/unauthorized.401.json` | 401 Unauthorized. |
| `errors/forbidden.403.json` | 403 Forbidden. |
| `errors/not_found.404.json` | 404 Not Found. |
| `errors/bad_request.400.json` | 400 Validation failure. |

## How to Use These Mocks

- For quick manual testing, use `curl --data @file.json` with the request bodies.
- For UI work, point your mock server (Mockoon, json-server, etc.) at these fixtures.
- Combine with the curl snippets in [api-overview](./api-overview.md) for copy/paste flows.
