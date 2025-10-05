# Quick Reference: Refactored Code Patterns

## 🔧 Configuration

### Access Configuration
```go
// Get global config instance
cfg := config.Get()

// Access specific settings
serverPort := cfg.Server.Port
dbDSN := cfg.Database.DSN()
jwtSecret := cfg.JWT.Secret
jwtTTL := cfg.JWT.DefaultTTL
```

### Environment Variables Required
```env
DB_USER=your_user
DB_PASSWORD=your_password
DB_HOST=localhost
DB_PORT=3306
DB_NAME=your_database
JWT_SECRET=your-secret-key
```

## 🎯 Handler Patterns

### Extract Actor from Context
```go
// Safe - panics if missing (use in authenticated routes)
actor := ActorFromContext(c)

// Safe - returns bool (use when actor is optional)
actor, ok := GetActorFromContext(c)
if !ok {
    return // Error already sent
}
```

### Pointer Conversions
```go
// Convert plan string to *models.Plan
planPtr := PlanPointer(planString)

// Convert *string while trimming/validating
strPtr := StringPointer(domainPtr)
```

### Error Handling in Handlers
```go
if err := httpx.HandleBindingError(c, err); err != nil {
    httpx.HandleError(c, err)
    return
}
```

## ✅ Validation Patterns

### Validate Tenant Input
```go
validated, err := ValidateTenantOnboardInput(req)
if err != nil {
    return fmt.Errorf("validation failed: %w", err)
}
// Use validated.Name, validated.Slug, etc.
```

### Validate User Input
```go
if err := ValidateUserInput(input); err != nil {
    return fmt.Errorf("invalid user input: %w", err)
}
```

### Validate Password Input
```go
if err := ValidatePasswordInput(input); err != nil {
    return fmt.Errorf("invalid password: %w", err)
}
```

## ❌ Error Handling

### Wrap Errors with Context
```go
// Always use %w to wrap errors
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// Add context to errors
if err := db.Create(&user).Error; err != nil {
    return fmt.Errorf("database error creating user %s: %w", user.Email, err)
}
```

### Never Ignore Errors
```go
// Bad ❌
db.Create(&user)

// Good ✅
if err := db.Create(&user).Error; err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}
```

## 🗂️ Constants

### Middleware Constants
```go
import "rtr-user-auth-service/middleware"

// HTTP Headers
headerValue := c.GetHeader(middleware.HeaderAuthorization)
c.Header(middleware.HeaderTenantID, tenantID)

// Context Keys
actor := c.MustGet(middleware.ContextKeyActor)
```

## 🔐 JWT Usage

### Sign JWT Token
```go
token, expiresAt, err := utils.SignJWT(
    userID, 
    tenantID, 
    email, 
    role, 
    cfg.JWT.DefaultTTL,
)
if err != nil {
    return fmt.Errorf("failed to sign JWT: %w", err)
}
```

## 🗄️ Database Patterns

### Initialize Database
```go
cfg := config.Get()
db := initializeDatabase(cfg)
// Connection pool is automatically configured
```

### Transaction Pattern
```go
err := db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }
    if err := tx.Create(&tenant).Error; err != nil {
        return fmt.Errorf("failed to create tenant: %w", err)
    }
    return nil
})
```

## 📝 Common Helper Functions

### handlers/helpers.go
```go
// Pointer conversions
PlanPointer(plan string) *models.Plan
StringPointer(s *string) *string

// Context extraction
ActorFromContext(c *gin.Context) services.UserRead
GetActorFromContext(c *gin.Context) (services.UserRead, bool)
```

## 🎨 Code Style Guidelines

### Variable Names
```go
// Short but meaningful
cfg := config.Get()
db := initializeDatabase(cfg)
req := &CreateUserRequest{}
```

### Error Handling
```go
// Error-first returns
func CreateUser(...) (*User, error) {
    if err := validate(); err != nil {
        return nil, err
    }
    // ... success path
    return user, nil
}
```

### Defer for Cleanup
```go
func process() error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel() // Always cleanup

    // ... processing
    return nil
}
```

## 🚫 What NOT to Do

### Don't Use Hardcoded Values
```go
// Bad ❌
secret := "my-secret"
port := "8080"

// Good ✅
cfg := config.Get()
secret := cfg.JWT.Secret
port := cfg.Server.Port
```

### Don't Use os.Getenv Directly
```go
// Bad ❌
secret := os.Getenv("JWT_SECRET")

// Good ✅
cfg := config.Get()
secret := cfg.JWT.Secret
```

### Don't Silently Fail
```go
// Bad ❌
if err != nil {
    return
}

// Good ✅
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

### Don't Use Magic Strings
```go
// Bad ❌
c.GetHeader("Authorization")
c.Set("actor", user)

// Good ✅
c.GetHeader(middleware.HeaderAuthorization)
c.Set(middleware.ContextKeyActor, user)
```

## 📦 Import Organization

```go
package mypackage

import (
    // Standard library first
    "context"
    "fmt"
    "strings"

    // External packages
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    // Internal packages
    "rtr-user-auth-service/config"
    "rtr-user-auth-service/models"
    "rtr-user-auth-service/services"
)
```

## 🧪 Testing Patterns

### Test with Config
```go
func TestSomething(t *testing.T) {
    cfg := &config.Config{
        JWT: config.JWTConfig{
            Secret: "test-secret",
            DefaultTTL: time.Hour,
        },
    }
    // Use cfg in test
}
```

### Mock Services
```go
type mockTenantService struct{}

func (m *mockTenantService) CreateTenant(...) error {
    return nil
}
```

## 🔗 Quick Links

- **Configuration:** `config/config.go`
- **Handler Helpers:** `handlers/helpers.go`
- **Validators:** `services/validators.go`
- **Constants:** `middleware/constants.go`
- **Full Guide:** `docs/REFACTORING_GUIDE.md`
- **Checklist:** `VERIFICATION_CHECKLIST.md`

## 💡 Pro Tips

1. **Always wrap errors** with `%w` for better error traces
2. **Use config package** instead of direct env access
3. **Extract validation** to `services/validators.go`
4. **Use helper functions** from `handlers/helpers.go`
5. **Add context to errors** for easier debugging
6. **Keep functions small** and focused on one thing
7. **Use constants** instead of magic strings
8. **Test with mocked config** for predictable tests

## ⚡ Performance Tips

1. **Reuse database connections** - Pool is configured
2. **Use prepared statements** for repeated queries
3. **Avoid unnecessary allocations** - Use helper functions
4. **Cache config access** - `cfg := config.Get()` once per function
5. **Use context timeouts** for long operations

---

**Quick Start:**
1. Copy `.env.example` to `.env`
2. Update database credentials
3. Set `JWT_SECRET`
4. Run `go build ./cmd/server`
5. Run `./main_refactored.exe`

**Need Help?** Check `docs/REFACTORING_GUIDE.md` for detailed examples.
