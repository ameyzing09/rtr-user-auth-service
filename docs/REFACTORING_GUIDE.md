# Refactoring Guide

This document explains the major refactoring changes made to improve code quality, maintainability, and adherence to Go best practices.

## Overview of Changes

The refactoring focused on:
1. **Eliminating hardcoded values** - Moved to centralized configuration
2. **Reducing code duplication** - Extracted common functions to helpers
3. **Improving error handling** - Added proper error wrapping with `%w`
4. **Following Go idioms** - Better package structure, defer usage, short variable names
5. **Enhancing performance** - Database connection pooling, reduced allocations
6. **Ensuring concurrency safety** - Proper use of contexts and thread-safe patterns

## Key Changes

### 1. Configuration Management

**New Package:** `config/config.go`

All environment variables and application settings are now managed through a centralized configuration system.

**Usage Example:**
```go
// Old way (scattered throughout code)
secret := os.Getenv("JWT_SECRET")
if secret == "" {
    secret = "dev-secret"
}

// New way (centralized)
cfg := config.Get()
secret := cfg.JWT.Secret
```

**Benefits:**
- Single source of truth for configuration
- Type-safe configuration access
- Built-in validation
- Default values managed in one place
- Easy to test with different configurations

### 2. Server Initialization

**File:** `cmd/server/main.go`

Extracted initialization logic into separate functions:
- `initializeDatabase(cfg)` - Database setup with pooling
- `initializeRouter(cfg, db)` - Router and middleware setup
- `startServer(cfg, router)` - HTTP server startup

**Benefits:**
- Cleaner main function
- Easier to test individual components
- Better separation of concerns
- Improved error handling

### 3. Handler Helpers

**New Package:** `handlers/helpers.go`

Common handler utilities extracted:
```go
// Pointer conversion
planPtr := helpers.PlanPointer(planString)
strPtr := helpers.StringPointer(optionalString)

// Context extraction
actor := helpers.ActorFromContext(c)
```

**Benefits:**
- Eliminates duplicate pointer conversion logic
- Consistent error handling
- Reusable across all handlers

### 4. Service Layer Validators

**New File:** `services/validators.go`

Validation logic extracted from service methods:
```go
validated, err := ValidateTenantOnboardInput(req)
if err != nil {
    return nil, err
}
```

**Benefits:**
- Separates validation from business logic
- Easier to test validation rules
- Consistent error messages
- Reusable validation functions

### 5. Middleware Constants

**New File:** `middleware/constants.go`

All middleware-related constants in one place:
```go
const (
    HeaderAuthorization  = "Authorization"
    HeaderTenantID       = "X-Tenant-ID"
    ContextKeyActor      = "actor"
    // ...
)
```

**Benefits:**
- No magic strings
- Type safety
- Easy to update
- Self-documenting code

## Migration Guide

### For Existing Code

#### 1. Update Main Function

```go
// Old
db := connectDatabase()
router := gin.Default()

// New
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

db := initializeDatabase(cfg)
router := initializeRouter(cfg, db)
```

#### 2. Update Environment Access

```go
// Old
port := os.Getenv("SERVER_PORT")
if port == "" {
    port = "8082"
}

// New
cfg := config.Get()
port := cfg.Server.Port
```

#### 3. Update Handler Code

```go
// Old
var planPtr *models.Plan
if req.Plan != "" {
    plan := models.Plan(req.Plan)
    planPtr = &plan
}

// New
planPtr := helpers.PlanPointer(req.Plan)
```

#### 4. Update Service Validation

```go
// Old
normalizedName := strings.TrimSpace(req.Name)
if normalizedName == "" {
    return ErrInvalidInput
}
adminName := strings.TrimSpace(req.AdminName)
if adminName == "" {
    return ErrInvalidInput
}
// ... more validation

// New
validated, err := ValidateTenantOnboardInput(req)
if err != nil {
    return err
}
```

### Environment Variables

Update your `.env` file with all required variables. See `.env.example` for a complete list.

**Required Variables:**
```env
# Database
DB_USER=your_user
DB_PASSWORD=your_password
DB_HOST=localhost
DB_PORT=3306
DB_NAME=your_database

# JWT
JWT_SECRET=your-secret-key
```

**Optional Variables with Defaults:**
```env
SERVER_PORT=8082
GIN_MODE=release
ENV=local
JWT_TTL=24h
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
```

## Best Practices

### 1. Error Handling

Always wrap errors with context using `%w`:

```go
// Good
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// Bad
if err != nil {
    return err
}
```

### 2. Configuration Access

Use `config.Get()` instead of direct `os.Getenv()`:

```go
// Good
cfg := config.Get()
secret := cfg.JWT.Secret

// Bad
secret := os.Getenv("JWT_SECRET")
```

### 3. Pointer Conversion

Use helper functions for pointer conversion:

```go
// Good
planPtr := helpers.PlanPointer(plan)

// Bad
var planPtr *models.Plan
if plan != "" {
    p := models.Plan(plan)
    planPtr = &p
}
```

### 4. Context Extraction

Use helper functions to extract values from context:

```go
// Good
actor := helpers.ActorFromContext(c)

// Bad
actor := c.MustGet("actor").(services.UserRead)
```

## Testing

### Unit Tests

Test configuration loading:
```go
func TestConfigLoad(t *testing.T) {
    os.Setenv("DB_USER", "testuser")
    cfg, err := config.Load()
    assert.NoError(t, err)
    assert.Equal(t, "testuser", cfg.Database.User)
}
```

Test validators:
```go
func TestValidateTenantInput(t *testing.T) {
    req := TenantOnboardAsyncRequest{
        Name: "Test Tenant",
        AdminName: "Admin",
        AdminEmail: "admin@test.com",
    }
    validated, err := ValidateTenantOnboardInput(req)
    assert.NoError(t, err)
    assert.Equal(t, "Test Tenant", validated.Name)
}
```

### Integration Tests

Test with full configuration:
```go
func TestIntegration(t *testing.T) {
    cfg := &config.Config{
        Database: config.DatabaseConfig{
            // test database config
        },
    }
    db := initializeDatabase(cfg)
    // ... rest of test
}
```

## Performance Improvements

### Database Connection Pooling

Now configured via environment variables:
```env
DB_MAX_OPEN_CONNS=25        # Maximum open connections
DB_MAX_IDLE_CONNS=5         # Idle connections in pool
DB_CONN_MAX_LIFETIME=5m     # Connection lifetime
```

### Reduced Allocations

- Helper functions optimize pointer conversions
- Config loaded once at startup
- Reusable validation functions

## Troubleshooting

### Config Not Initialized Error

If you see "config not initialized", ensure `config.Load()` is called before any code that uses `config.Get()`:

```go
func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }
    
    // Now config.Get() will work
}
```

### Missing Environment Variables

Check `.env` file exists and contains required variables. The application will fail to start if required database variables are missing.

### Import Issues

Update imports after refactoring:
```go
import (
    "rtr-user-auth-service/config"
    "rtr-user-auth-service/handlers"
    // ...
)
```

## Future Improvements

See `REFACTORING_SUMMARY.md` for:
- Additional refactoring opportunities
- Performance optimization suggestions
- Architecture improvements
- Testing recommendations

## Questions or Issues?

If you encounter any issues or have questions about the refactoring:
1. Check this guide and `REFACTORING_SUMMARY.md`
2. Review the example code in the repository
3. Check existing tests for usage patterns
4. Consult with the team

## Conclusion

These refactoring changes significantly improve:
- Code maintainability
- Type safety
- Error handling
- Performance
- Testability

Follow the patterns established in this refactoring for any new code to maintain consistency and quality.
