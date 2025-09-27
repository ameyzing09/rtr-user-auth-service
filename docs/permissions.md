# Server-Side Permission System

This document describes the server-side permission enforcement system implemented in the rtr-user-auth-service.

## Overview

The system enforces permissions at the server level, making the backend the source of truth for access control. While the UI may hide buttons or features, the server validates all requests and enforces proper authorization.

## Architecture

### 1. Role-Based Access Control (RBAC)

The system uses a hierarchical role-based access control model with the following roles:

- **SUPERADMIN**: Full system access, can manage tenants and bypass tenant boundaries on control-plane routes only
- **ADMIN**: Tenant-level administrator, can manage users and tenant settings
- **HR**: Human resources role, can list and create users, view tenant settings
- **INTERVIEWER**: Limited access, can only manage their own profile
- **CANDIDATE**: Limited access, can only manage their own profile

### 2. Components

#### Role Enum (`models/role.go`)
```go
type Role string

const (
    RoleSuperAdmin  Role = "SUPERADMIN"
    RoleAdmin       Role = "ADMIN"
    RoleHR          Role = "HR"
    RoleInterviewer Role = "INTERVIEWER"
    RoleCandidate   Role = "CANDIDATE"
)
```

#### Actor Type (`services/contracts.go`)
```go
type UserRead struct {
    ID       string
    TenantID string
    Email    string
    Role     models.Role
}
```

#### Authentication Middleware (`middleware/auth.go`)
- Parses JWT tokens and extracts user information
- Builds `services.UserRead` actor object
- Enforces tenant boundary checks (except for SUPERADMIN)
- Sets actor in request context

#### Role Gates Middleware (`middleware/roles.go`)
- `RequireRole(role)`: Requires a specific role
- `RequireAny(roles...)`: Requires any of the specified roles

#### Policy System (`policy/policy.go`)
- Action-based permission checking
- Fine-grained control for specific operations
- Extensible for complex permission scenarios

## Permission Matrix

| Action | SUPERADMIN | ADMIN | HR | INTERVIEWER | CANDIDATE |
|--------|------------|-------|----|-----------|-----------| 
| **User Management** |
| List users | ✅ | ✅ | ✅ | ❌ | ❌ |
| Create users | ✅ | ✅ | ✅ | ❌ | ❌ |
| Update users | ✅ | ✅ | ❌ | ❌ | ❌ |
| Delete users | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Tenant Management** |
| Create tenants | ✅ | ❌ | ❌ | ❌ | ❌ |
| Update tenants | ✅ | ❌ | ❌ | ❌ | ❌ |
| Delete tenants | ✅ | ❌ | ❌ | ❌ | ❌ |
| View tenants | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Tenant Settings** |
| View settings | ✅ | ✅ | ✅ | ❌ | ❌ |
| Update settings | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Profile Management** |
| View profile | ✅ | ✅ | ✅ | ✅ | ✅ |
| Update profile | ✅ | ✅ | ✅ | ✅ | ✅ |

## Route Protection

### Public Routes (No Authentication)
- `POST /login` - User authentication
- `GET /tenant/settings` - Public tenant settings view

### Protected Routes (Tenant Context + Authentication)
- `GET /me` - User profile (all authenticated users)
- `POST /me/change-password` - Change password (all authenticated users)
- `GET /users` - List users (ADMIN, HR)
- `POST /users` - Create users (ADMIN, HR)
- `PUT /tenant/settings` - Update tenant settings (ADMIN only)

### Admin Routes (SUPERADMIN Only, No Tenant Context)
- `POST /tenant/create` - Create new tenant
- `GET /tenant/:id` - Get tenant details
- `GET /tenant/:id/status` - Get tenant status
- `POST /tenant/:id/retry` - Retry tenant provisioning

## Implementation Examples

### Using Role Gates in Routes

```go
// Require specific role
admin.POST("/tenant/create", middleware.RequireRole(models.RoleSuperAdmin), handler.Create)

// Require any of multiple roles
protected.GET("/users", middleware.RequireAny(models.RoleAdmin, models.RoleHR), handler.ListUsers)
```

### Using Policy System in Handlers

```go
func (h *UserHandler) DeleteUser(c *gin.Context) {
    actor, _ := c.Get("actor").(services.UserRead)
    
    if !policy.Can(actor.Role, policy.ActionUserDelete) {
        c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
        return
    }
    
    // Proceed with deletion
}
```

### Checking Multiple Actions

```go
// Check if user can perform any of the actions
if policy.CanAny(actor.Role, policy.ActionUserList, policy.ActionUserCreate) {
    // User can list or create users
}

// Check if user can perform all actions
if policy.CanAll(actor.Role, policy.ActionUserList, policy.ActionUserCreate) {
    // User can both list and create users
}
```

## Security Features

### 1. Tenant Boundary Enforcement

#### Control-Plane Routes (`/admin/*`)
- **SUPERADMIN bypass applies**: SUPERADMIN can access these routes without tenant context
- **Purpose**: System administration, tenant management, cross-tenant operations
- **Examples**: `POST /tenant/create`, `GET /tenant/:id`, `POST /tenant/:id/retry`

#### Tenant-Scoped Routes (all other protected routes)
- **SUPERADMIN bypass does NOT apply**: Even SUPERADMIN must have matching tenant context
- **Purpose**: Prevents accidental cross-tenant access and maintains data isolation
- **Examples**: `GET /users`, `PUT /tenant/settings`, `GET /me`
- **Security**: Prevents someone with SUPERADMIN token from directly hitting tenant routes

#### Cross-Tenant Access Prevention
- Users can only access resources within their tenant
- Cross-tenant access is automatically blocked on all tenant-scoped routes
- This prevents data leakage and maintains tenant isolation

### 2. JWT Token Validation
- All protected routes require valid JWT tokens
- Token expiration is enforced
- Role information is extracted from token claims

### 3. Request Context Validation
- Actor information is validated on every request
- Invalid or missing actor context results in 401 Unauthorized
- Role mismatches result in 403 Forbidden

## Error Responses

### 401 Unauthorized
```json
{
  "error": "Missing or invalid Authorization header"
}
```

### 403 Forbidden
```json
{
  "error": "forbidden"
}
```

### 403 Forbidden (Tenant Mismatch)
```json
{
  "error": "Access to this tenant is forbidden"
}
```

## Testing

The permission system includes comprehensive tests:

### Unit Tests
- **Role gate middleware tests** (`middleware/roles_test.go`): Test individual role gate functions
- **Policy system tests** (`policy/policy_test.go`): Test action-based permission logic

### Integration Tests
- **Auth middleware integration tests** (`middleware/auth_integration_test.go`): Test tenant boundary enforcement
- **Role gates integration tests** (`middleware/roles_integration_test.go`): Test role gates with tenant context

#### Critical Integration Test Scenarios

The integration tests verify these critical security invariants:

1. **Tenant Mismatch Prevention**:
   ```go
   // HR from tenant A cannot access tenant B's API
   TestAuthMiddleware_TenantMismatch/HR_from_tenant_A_cannot_access_tenant_B
   ```

2. **SUPERADMIN Bypass Behavior**:
   ```go
   // SUPERADMIN can access any tenant (tenant boundary bypass)
   TestAuthMiddleware_TenantMismatch/SUPERADMIN_can_access_any_tenant
   
   // SUPERADMIN can access any role-gated route
   TestRoleGates_WithTenantContext/SUPERADMIN_can_access_ADMIN_route_in_any_tenant
   ```

3. **Role Gate Enforcement**:
   ```go
   // ADMIN cannot access routes requiring different roles
   TestRoleGates_WithTenantContext/HR_cannot_access_ADMIN-only_route
   ```

4. **Cross-Tenant Access Blocking**:
   ```go
   // Regular users cannot cross tenant boundaries
   TestRoleGates_WithTenantContext/ADMIN_cannot_access_route_in_different_tenant
   ```

### Running Tests

```bash
# Run all permission system tests
go test ./middleware/... ./policy/...

# Run specific integration tests
go test ./middleware/... -v -run="TestAuthMiddleware_TenantMismatch"
go test ./middleware/... -v -run="TestRoleGates_WithTenantContext"

# Run with coverage
go test ./middleware/... ./policy/... -cover
```

### Test Coverage

The integration tests provide comprehensive coverage of:
- ✅ JWT token validation and parsing
- ✅ Tenant boundary enforcement
- ✅ Role-based access control
- ✅ SUPERADMIN bypass behavior
- ✅ Cross-tenant access prevention
- ✅ Error handling and response codes
- ✅ Actor context validation

## Best Practices

1. **Always use middleware for route protection** - Don't rely on handler-level checks alone
2. **Use policy system for complex logic** - When you need fine-grained control
3. **Test permission scenarios** - Ensure all role combinations work correctly
4. **Document permission changes** - Update this document when adding new roles or permissions
5. **Principle of least privilege** - Grant minimum required permissions

### Permission Review Process

When introducing new endpoints or features, follow this review process to prevent permission drift:

1. **Identify required actions** - What specific operations does the endpoint perform?
2. **Map to existing actions** - Can you reuse existing policy actions?
3. **Define new actions** - If needed, add new actions to the policy system
4. **Update permission matrix** - Document which roles can perform the new actions
5. **Add role gates** - Apply appropriate middleware to routes
6. **Write tests** - Test all role combinations and edge cases
7. **Review with team** - Ensure the permission model makes sense
8. **Update documentation** - Keep this document current

### Common Permission Anti-Patterns

❌ **Don't do this:**
```go
// Manual role checking in handlers
if actor.Role == "ADMIN" || actor.Role == "HR" {
    // Allow access
}
```

✅ **Do this instead:**
```go
// Use middleware for route protection
protected.GET("/users", middleware.RequireAny(models.RoleAdmin, models.RoleHR), handler.ListUsers)

// Use policy system for complex logic
if !policy.Can(actor.Role, policy.ActionUserList) {
    return 403 Forbidden
}
```

### Multi-Service Considerations

In multi-service architectures, permission drift can happen easily. To prevent this:

1. **Centralize permission logic** - Use shared policy packages
2. **Version permission APIs** - When changing permissions, version the changes
3. **Document breaking changes** - Clearly communicate permission changes
4. **Test integration scenarios** - Verify cross-service permission enforcement
5. **Monitor permission failures** - Track 403 errors to identify drift

## Adding New Permissions

1. **Define new actions** in `policy/policy.go`:
   ```go
   const (
       ActionNewFeature Action = "new_feature:action"
   )
   ```

2. **Update permission matrix** in `policy.Can()` function

3. **Add role gates** to routes as needed

4. **Update tests** to cover new permission scenarios

5. **Update documentation** with new permission matrix

## Migration Notes

### Database Schema Changes

#### Migration `004_add_superadmin_role.up.sql`
- Added `SUPERADMIN` to the `users.role` ENUM
- Updated role constraints to include the new role

#### Migration `005_create_tenants_async.up.sql`
- Added `plan` ENUM: `BASIC`, `STARTER`, `GROWTH`, `ENTERPRISE`, `ON_PREM`
- Added `status` ENUM: `PENDING`, `PROVISIONING`, `AWAITING_BRANDING`, `ACTIVE`, `FAILED`, `SUSPENDED`, `DELETED`
- Added new columns to `tenants` table: `domain`, `slug`, `plan`, `status`, `created_by`, `failed_reason`
- Added `is_owner` boolean to `users` table

#### Migration `006_outbox_idempotency.up.sql`
- Created `outbox` table for event sourcing
- Created `idempotency_keys` table for request deduplication

### Application Changes

- **Existing users**: Retain their current roles, no automatic role changes
- **New tenant creation**: Requires SUPERADMIN role (enforced by middleware)
- **Tenant boundary enforcement**: Automatic for all non-SUPERADMIN users
- **Role validation**: All role checks now use the updated ENUM values
- **Permission system**: New middleware and policy system replaces manual checks

### Breaking Changes

- **API endpoints**: New admin routes require SUPERADMIN role
- **Role validation**: Stricter role checking with explicit ENUM values
- **Tenant isolation**: Enhanced tenant boundary enforcement
- **Error responses**: Standardized error codes and messages

### Deployment Considerations

1. **Database migration**: Run migrations in order (004, 005, 006)
2. **Role assignment**: Manually assign SUPERADMIN role to system administrators
3. **API clients**: Update to handle new error codes and permission requirements
4. **Testing**: Verify all role combinations work correctly after deployment
