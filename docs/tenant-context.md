# Tenant Context Header Contract

This document describes the headers the user-auth service expects on tenant-scoped API calls. Share it with any client or proxy that needs to call routes protected by `TenantContext` middleware.

## Overview

`TenantContext` enforces that requests identify the tenant through signed headers. The signature prevents callers from guessing or spoofing another tenant id and the timestamp blocks replay attempts. Clients must generate all headers atomically; missing or stale values result in a `401`.

## Header Summary

| Header | Required | Notes |
| --- | --- | --- |
| `X-Tenant-Id` | Yes | Tenant UUID from provisioning or login response. |
| `X-Tenant-Domain` | Optional | Match against tenant domain; include when the caller knows the canonical domain. |
| `X-Tenant-Ts` | Yes | UTC Unix timestamp divided by 60 (integer minutes). Valid for +/-2 minutes. |
| `X-Tenant-Sig` | Yes | Base64url-encoded HMAC-SHA256 of `<tenantId>.<domain>.<ts>` using `TENANT_CTX_SECRET`. |

The payload always includes two dots. When no domain is supplied the value looks like `tenant-123..28934721`.

## Timestamp Rules (`X-Tenant-Ts`)

- Compute as `Math.floor(Date.now() / 1000 / 60)` (JavaScript) or `time.Now().UTC().Unix() / 60` (Go).
- Use UTC only. Local timezones will drift and fail validation.
- Each signature is valid for the minute it was created plus a 2 minute skew window. Regenerate on every request or retry.

## Signature Rules (`X-Tenant-Sig`)

1. Build the payload string: `<tenantId>.<domain or empty string>.<ts>`.
2. Sign with HMAC-SHA256 using the current `TENANT_CTX_SECRET` (rotate support: server also accepts `TENANT_CTX_SECRET_PREV`).
3. Base64url-encode the digest and strip padding (`=`). Replace `+` with `-` and `/` with `_`.
4. Ship the signature alongside the source headers.

Never embed `TENANT_CTX_SECRET` in browser JavaScript. Instead, have the UI call a backend-for-frontend (BFF) or API gateway that holds the secret and forwards requests with the signed headers.

## Reference Implementation (Node.js/TypeScript)

```ts
import crypto from "node:crypto";

export function makeTenantContextHeaders(
  tenantId: string,
  domain: string | undefined,
  secret: string
) {
  const ts = Math.floor(Date.now() / 1000 / 60).toString();
  const payload = `${tenantId}.${domain ?? ""}.${ts}`;
  const sig = crypto
    .createHmac("sha256", secret)
    .update(payload, "utf8")
    .digest("base64")
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");

  return {
    "X-Tenant-Id": tenantId,
    ...(domain ? { "X-Tenant-Domain": domain } : {}),
    "X-Tenant-Ts": ts,
    "X-Tenant-Sig": sig,
  };
}
```

Use the helper inside your server proxy just before making the outbound call to the auth service.

## Example Request (curl)

```bash
TENANT_ID="tenant-123"
TENANT_DOMAIN="widgets.example.com"
SECRET="$TENANT_CTX_SECRET"
TS=$(($(date -u +%s) / 60))
PAYLOAD="${TENANT_ID}.${TENANT_DOMAIN}.${TS}"
SIG=$(python3 - <<'PY'
import base64, hashlib, hmac, os
payload = os.environ["PAYLOAD"].encode()
secret = os.environ["SECRET"].encode()
print(base64.urlsafe_b64encode(hmac.new(secret, payload, hashlib.sha256).digest()).decode().rstrip('='))
PY
)

curl -X GET "$AUTH_BASE_URL/v1/me" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-Id: $TENANT_ID" \
  -H "X-Tenant-Domain: $TENANT_DOMAIN" \
  -H "X-Tenant-Ts: $TS" \
  -H "X-Tenant-Sig: $SIG"
```

If you omit the domain header, leave the middle segment empty when building `PAYLOAD` and drop the header from the curl command.

## Failure Modes

- Missing headers -> `401` with `{ "error": "missing tenant signature headers" }`.
- Non-numeric timestamp -> `401` with `{ "error": "invalid tenant timestamp" }`.
- Timestamp older than 2 minutes -> `401` with `{ "error": "tenant context expired" }`.
- Signature mismatch -> `401` with `{ "error": "invalid tenant signature" }`.
- Tenant id not found -> `404` with `{ "error": "tenant not found" }`.
- Domain mismatch (when provided) -> `403` with `{ "error": "tenant domain mismatch" }`.

## Integration Checklist

- [ ] Store `TENANT_CTX_SECRET` in backend config management (env var, vault, etc.).
- [ ] Regenerate headers per outbound request or retry.
- [ ] Include `X-Tenant-Domain` when the client routes per-domain traffic.
- [ ] Monitor `401`/`403` errors to detect clock skew or stale secrets.
- [ ] Refresh cache after rotating `TENANT_CTX_SECRET`; new signatures take effect immediately.

For further questions, contact the Auth Platform team or check `middleware/tenant_context.go` for the definitive implementation.
