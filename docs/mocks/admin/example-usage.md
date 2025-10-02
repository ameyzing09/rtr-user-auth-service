# Superadmin Workflow Examples

All commands assume `BASE_URL=http://localhost:8082` and a valid SUPERADMIN bearer token stored in `$TOKEN`.

## Login
```bash
curl -X POST "$BASE_URL/admin/login" \
  -H "Content-Type: application/json" \
  --data @docs/mocks/login.request.json
```
Response includes `PlatformBranding` for rendering nav/sidebars (see `admin/login.response.200.json`).

## List Tenants
```bash
curl -X GET "$BASE_URL/admin/tenants" \
  -H "Authorization: Bearer $TOKEN"
```
Returns `admin/tenants.list.response.200.json`.

## Create Tenant (Async)
```bash
curl -X POST "$BASE_URL/tenant/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Idempotency-Key: $(uuidgen | tr 'A-Z' 'a-z')" \
  -H "Content-Type: application/json" \
  --data @docs/mocks/admin/tenants.onboard.request.json
```
- 202 response ? creation kicked off.
- Repeat the same request with the same `Idempotency-Key` to receive the cached 200 payload (`tenants.create.response.200.cached.json`).

## Tenant Details & Status
```bash
curl -X GET "$BASE_URL/tenant/tnt-0000-1111-2222-333344445555" \
  -H "Authorization: Bearer $TOKEN"

curl -X GET "$BASE_URL/tenant/tnt-0000-1111-2222-333344445555/status" \
  -H "Authorization: Bearer $TOKEN"
```
Use the mock files `tenants.get.*` and `tenants.status.*` for UI states (active, failed, etc.).

## Retry Provisioning
```bash
curl -X POST "$BASE_URL/tenant/tnt-0000-1111-2222-333344445555/retry" \
  -H "Authorization: Bearer $TOKEN"
```
Returns `202 Accepted` with an empty body.

## Logout
```bash
curl -X POST "$BASE_URL/admin/logout" \
  -H "Authorization: Bearer $TOKEN"
```
