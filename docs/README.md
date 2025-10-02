# RTR User Auth Service Docs

This folder contains everything you need to work with the `rtr-user-auth-service`, including API routes, sample payloads, and operational guidance. Each sub-document is short and task oriented?start here and jump to whatever you need.

## Quick Map

| Topic | File |
| --- | --- |
| HTTP endpoints, auth model, curl cheatsheet | [api-overview](./api-overview.md) |
| Ready-to-use JSON payloads | [mock-responses](./mock-responses.md) |
| Role matrix and permission rules | [permissions](./permissions.md) |
| Tenant slug creation rules | [slug-configuration](./slug-configuration.md) |
| Logging levels & env knobs | [logging](./logging.md) |
| Admin control plane payloads | [mocks/admin/README.md](./mocks/admin/README.md) |

## Getting Started

1. Run the service locally (see project README for build instructions).
2. Acquire or mint a JWT; SUPERADMIN tokens unlock control-plane endpoints.
3. Use the curl commands in `api-overview.md` to poke the API.
4. If you are mocking the backend, point your mock tool to the JSON bodies in `docs/mocks/`.

Happy shipping! If something is missing, extend these files?they?re intentionally short and easy to edit.
