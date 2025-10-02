# Tenant Slug Rules

Slugs uniquely identify tenants in URLs and infrastructure names. They are generated automatically during onboarding but can be supplied manually.

## Validation

- Lowercase letters, numbers, and single dashes only.
- Must start and end with an alphanumeric character.
- Length: 3 to 30 characters.
- Regex: `^[a-z0-9](?:[a-z0-9-]*[a-z0-9])$`.

## Generation Flow

1. Normalize company name (trim, lowercase, spaces to dashes).
2. Deduplicate consecutive dashes.
3. If slug already exists, suggest alternatives:
   - `<slug>-hq`
   - `<slug>-io`
   - `<slug>-team`
4. Persist the final slug in `tenants.slug` (unique index).

## API Behaviour

- POST `/tenant/create` accepts optional `slug` via name (implicit).
- On conflict the API returns HTTP 409 with `TENANT_SLUG_TAKEN` and `suggestions` array.
- Slug is returned on create, list, and get endpoints.

## Tips

- Keep slugs stable; other services cache them.
- Avoid manual edits unless you coordinate across the platform.
- When migrating existing tenants, pre-fill `slug` to preserve vanity URLs.
