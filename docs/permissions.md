# Permission Matrix

The service uses role-based access control. Roles are immutable strings embedded in JWTs.

## Roles

| Role | Scope |
| --- | --- |
| `SUPERADMIN` | Control plane (all tenants). |
| `ADMIN` | Tenant admin. Manages users and settings. |
| `HR` | Tenant HR manager. Can manage users (no settings). |
| `INTERVIEWER` | Read-only profile access. |
| `CANDIDATE` | Self-service endpoints only. |

## Endpoint Access

| Endpoint | Roles |
| --- | --- |
| `POST /login` | Public |
| `GET /me` | Any authenticated role |
| `POST /me/change-password` | Any authenticated role |
| `POST /logout` | Any authenticated role |
| `GET /users`, `POST /users` | `ADMIN`, `HR` |
| `GET /tenant/settings`, `PUT /tenant/settings` | `ADMIN` (PUT only) |
| `/admin/*` routes | `SUPERADMIN` |

## Claims in JWTs

```
{
  "uid": "u-...",
  "tid": "tnt-...",
  "role": "ADMIN",
  "email": "user@tenant.com",
  "exp": 1759044368
}
```

## Adding Tenants for Local Dev

1. Insert a tenant row in the database.
2. Seed a user with `role='SUPERADMIN'` or higher.
3. Issue a JWT signed with `JWT_SECRET` pointing to that user.

No runtime role changes are supported?mint a new token when promoting users.
