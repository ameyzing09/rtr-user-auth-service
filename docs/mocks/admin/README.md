# Control Plane Mock Payloads

These fixtures power the SUPERADMIN UI and integration tests. All paths are relative to `docs/mocks/admin/`.

## Endpoints & Payloads

| Endpoint | Method | Fixture |
| --- | --- | --- |
| `/admin/login` | POST | `login.response.200.json` |
| `/admin/logout` | POST | *(returns 204, no body)* |
| `/admin/tenants` | GET | `tenants.list.response.200.json` |
| `/tenant/create` | POST | `tenants.onboard.request.json`, `tenants.onboard.response.202.json`, `tenants.create.response.200.cached.json` |
| `/tenant/:id` | GET | `tenants.get.response.200.json`, `tenants.get.active.response.200.json`, `tenants.get.failed.response.200.json` |
| `/tenant/:id/status` | GET | `tenants.status.*.json` |
| `/tenant/:id/retry` | POST | *(returns 202, no body)* |

## Error Fixtures

| Scenario | File |
| --- | --- |
| Missing SUPERADMIN role | `tenants.create.error.403.superadmin.json` |
| Domain already in use | `tenants.create.error.409.domain.json` |
| Slug taken | `tenants.create.error.409.slug.json` |
| Idempotency key reused | `tenants.create.error.409.idempotency.json` |
| Missing Idempotency key | `tenants.create.error.422.no_idempotency.json` |
| Validation failures | `tenants.create.error.400.validation.json` |
| Tenant not found | `tenants.get.error.404.notfound.json` |

## Using the Files

- Every mock follows the live casing (`Token`, `PlatformBranding`, etc.) to keep UI bindings simple.
- Control plane errors always return uppercase `code` values?bubble them to the UI.
- When scripting curl, pair these payloads with the snippets in [`docs/api-overview.md`](../api-overview.md).

Need a variant that isn?t here? Add a new JSON file and reference it in this table.
