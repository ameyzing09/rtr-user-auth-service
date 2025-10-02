# API Overview

All endpoints use JSON and standard HTTP codes. Authentication is via JWT bearer tokens. Tenant-scoped routes require both a valid token and the `X-Tenant-ID` header supplied by login. Control-plane routes require a SUPERADMIN token.

Base URL examples:
- Local dev: `http://localhost:8082`
- Production: `https://auth.<env>.recrutr.in`

## Public Routes

| Method | Path | Description |
| --- | --- | --- |
| POST | `/login` | Exchange email/password for a tenant-scoped JWT. Returns `X-Tenant-ID` header. |

**Curl**
```bash
curl -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  --data @docs/mocks/login.request.json
```

## Tenant Routes (JWT + `X-Tenant-ID`)

| Method | Path | Notes |
| --- | --- | --- |
| GET | `/me` | Current user profile. |
| POST | `/me/change-password` | Requires `current_password` and `new_password`. |
| POST | `/logout` | Clears client-side session (stateless). |
| GET | `/users` | ADMIN or HR. Lists tenant users. |
| POST | `/users` | ADMIN or HR. Creates user, returns `temporary_password`. |
| GET | `/tenant/settings` | Tenant configuration blob. |
| PUT | `/tenant/settings` | ADMIN only. Replaces tenant configuration. |

Headers:
```
Authorization: Bearer <token>
X-Tenant-ID: <tenant-id>
```

## Control-Plane Routes (SUPERADMIN)

| Method | Path | Description |
| --- | --- | --- |
| POST | `/admin/login` | Returns SUPERADMIN JWT plus `PlatformBranding`. |
| POST | `/admin/logout` | Clears control-plane session cache. |
| GET | `/admin/tenants` | Lists every tenant with plan, status, timestamps. |
| POST | `/tenant/create` | Async tenant onboarding. Requires `Idempotency-Key`. |
| GET | `/tenant/:id` | Fetch tenant metadata. |
| GET | `/tenant/:id/status` | Provisioning status timeline. |
| POST | `/tenant/:id/retry` | Requeues provisioning (idempotent). |

Headers:
```
Authorization: Bearer <superadmin-token>
Idempotency-Key: <uuid>   # create only
```

## Responses & Errors

- Success payloads are camelCase/TitleCase to match existing clients.
- Tenant errors return `{ "error": "message" }`.
- Control-plane errors return `{ "code": "ERR_CODE", "message": "human readable" }` and may include `suggestions`.

For concrete JSON bodies reference [mock-responses](./mock-responses.md) and the files under `docs/mocks/`.


> Dev mode: set `SUPERADMIN_DEV_TOKEN` (default `dev-superadmin`) when `ENV` is `local` or `dev` to exercise admin routes without issuing a JWT. Never enable this in higher environments.
