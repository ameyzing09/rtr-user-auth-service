# Superadmin Change Password Endpoint

## Overview
This endpoint allows superadmin users to reset the password of any user who has `force_password_reset` set to `true`. The system automatically generates a new temporary password, similar to the tenant creation process. This is useful for administrative password resets.

## Endpoint
```
POST /admin/change-password
```

## Authentication
- Requires superadmin role (`SUPERADMIN`)
- Must be called with control plane scope (no tenant context)
- Requires valid JWT token in Authorization header

## Request Body
```json
{
  "user_id": "string (required)",
  "tenant_id": "string (required)"
}
```

## Response
- **200 OK**: Password reset successfully
  ```json
  {
    "temporary_password": "generated-temp-password"
  }
  ```
- **400 Bad Request**: Invalid request body or validation errors
- **401 Unauthorized**: Missing or invalid authentication
- **403 Forbidden**: 
  - User is not superadmin
  - Target user does not have `force_password_reset = true`
- **404 Not Found**: Target user not found

## Security Notes
- Only works when target user has `force_password_reset = true`
- Automatically generates a secure temporary password using `utils.GenerateTempPassword()`
- Automatically clears the `force_password_reset` flag after successful password change
- Password is hashed before storage
- Requires superadmin privileges
- Generated password is returned in the response for immediate use

## Example Usage
```bash
curl -X POST http://localhost:8080/admin/change-password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <superadmin-jwt-token>" \
  -d '{
    "user_id": "user-123",
    "tenant_id": "tenant-456"
  }'
```

## Response Example
```json
{
  "temporary_password": "TempPass123!@#"
}
```

## Implementation Details
- Handler: `UserHandler.SuperadminChangePassword`
- Service: `AuthService.SuperadminChangePassword`
- Repository: Uses existing `UpdatePassword` method with `clearForce = true`
- Middleware: `ControlPlaneScope`, `AuthMiddleware`, `RequireRole(SUPERADMIN)`
- Password Generation: Uses `utils.GenerateTempPassword()` (same as tenant creation)
