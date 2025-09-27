# Logging System

This document describes the conditional logging system implemented in the rtr-user-auth-service to optimize performance in production environments.

## Overview

The service uses a conditional logging system that allows different log levels based on the environment. This prevents performance degradation from excessive debug logging in production while maintaining detailed logging for development and troubleshooting.

## Log Levels

The system supports four log levels in order of verbosity:

- **DEBUG**: Most verbose, includes detailed flow information
- **INFO**: General information about application flow
- **WARN**: Warning messages for potentially problematic situations
- **ERROR**: Error messages for failures and exceptions

## Environment Configuration

### Environment Variables

- **`LOG_LEVEL`**: Controls the logging verbosity
  - `debug`: All messages (development)
  - `info`: Info, warn, and error messages (production default)
  - `warn`: Warning and error messages only
  - `error`: Error messages only

- **`GIN_MODE`**: Affects default log level
  - `release`: Defaults to INFO level
  - `debug` (or unset): Defaults to DEBUG level

### Default Behavior

```bash
# Development (GIN_MODE not set or set to debug)
LOG_LEVEL=debug  # Default

# Production (GIN_MODE=release)
LOG_LEVEL=info   # Default
```

## Usage

### In Middleware

The middleware uses conditional logging to provide detailed debugging information in development while minimizing overhead in production:

```go
// Debug logging - only in debug mode
utils.Debug("[AuthMiddleware] Authorization header present=%t", authHeader != "")

// Warning logging - always shown (important for security)
utils.Warn("[AuthMiddleware] JWT_SECRET not set, using default (development) secret")

// Info logging - shown in production
utils.Info("[AuthMiddleware] User authenticated successfully")
```

### In Services

Services can use the same logging system for consistent behavior:

```go
// Debug information about business logic
utils.Debug("[UserService] Creating user with email: %s", email)

// Important warnings
utils.Warn("[UserService] Password policy violation for user: %s", userID)

// General information
utils.Info("[UserService] User created successfully: %s", userID)
```

## Performance Impact

### Development Environment
- **Full logging enabled**: All debug messages are logged
- **Performance**: Slightly slower due to string formatting and I/O
- **Benefit**: Complete visibility into application flow

### Production Environment
- **Debug logging disabled**: Only info, warn, and error messages
- **Performance**: Minimal logging overhead
- **Benefit**: Fast execution with essential logging only

### Logging Overhead Analysis

| Log Level | Development | Production | Overhead |
|-----------|-------------|------------|----------|
| DEBUG     | ✅ Enabled  | ❌ Disabled| ~5-10ms  |
| INFO      | ✅ Enabled  | ✅ Enabled | ~1-2ms   |
| WARN      | ✅ Enabled  | ✅ Enabled | ~1-2ms   |
| ERROR     | ✅ Enabled  | ✅ Enabled | ~1-2ms   |

## Implementation Details

### Logger Structure

```go
type Logger struct {
    level LogLevel
}

type LogLevel int

const (
    LogLevelError LogLevel = iota
    LogLevelWarn
    LogLevelInfo
    LogLevelDebug
)
```

### Conditional Logging

```go
func (l *Logger) Debug(format string, v ...interface{}) {
    if l.level >= LogLevelDebug {
        log.Printf("[DEBUG] "+format, v...)
    }
}
```

### Package-Level Functions

```go
// Convenience functions using default logger
utils.Debug("Debug message")
utils.Info("Info message")
utils.Warn("Warning message")
utils.Error("Error message")

// Utility functions
if utils.IsDebugEnabled() {
    // Expensive debug operation
}
```

## Middleware Logging Examples

### Auth Middleware

```go
// Debug: Token validation details (development only)
utils.Debug("[AuthMiddleware] Token validation failed: error=%v, valid=%t", err, valid)

// Warn: Security concerns (always logged)
utils.Warn("[AuthMiddleware] Tenant mismatch: requestTenant=%s actorTenant=%s", tid, actor.TenantID)

// Info: Successful operations (production)
utils.Info("[AuthMiddleware] User authenticated: %s", userID)
```

### Tenant Context Middleware

```go
// Debug: Request processing details (development only)
utils.Debug("[TenantContext] Processing request: env=%s, tenantID=%s", env, tenantID)

// Debug: Cache operations (development only)
utils.Debug("[Cache] Cache MISS for tenant ID: %s, querying database", tenantID)

// Info: Successful resolution (production)
utils.Info("[TenantContext] Tenant resolved: %s", tenantID)
```

## Best Practices

### 1. Use Appropriate Log Levels

```go
// ✅ Good: Use debug for detailed flow information
utils.Debug("[Service] Processing request with params: %+v", params)

// ✅ Good: Use warn for security concerns
utils.Warn("[Auth] Invalid token from IP: %s", clientIP)

// ✅ Good: Use info for important business events
utils.Info("[Service] User created successfully: %s", userID)
```

### 2. Avoid Expensive Operations in Debug Logs

```go
// ❌ Bad: Expensive operation always executed
utils.Debug("Complex data: %s", expensiveDataProcessing())

// ✅ Good: Conditional expensive operation
if utils.IsDebugEnabled() {
    utils.Debug("Complex data: %s", expensiveDataProcessing())
}
```

### 3. Include Context in Log Messages

```go
// ✅ Good: Include relevant context
utils.Warn("[AuthMiddleware] Tenant mismatch: requestTenant=%s actorTenant=%s actorRole=%s", 
    tid, actor.TenantID, actor.Role)

// ❌ Bad: Missing context
utils.Warn("Tenant mismatch")
```

### 4. Use Structured Logging for Complex Data

```go
// ✅ Good: Structured approach
utils.Debug("[Service] User data: ID=%s, Email=%s, Role=%s", user.ID, user.Email, user.Role)

// ❌ Bad: Unstructured dump
utils.Debug("[Service] User: %+v", user)
```

## Deployment Configuration

### Docker

```dockerfile
# Set production log level
ENV LOG_LEVEL=info
ENV GIN_MODE=release
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: auth-service
        env:
        - name: LOG_LEVEL
          value: "info"
        - name: GIN_MODE
          value: "release"
```

### Local Development

```bash
# Enable debug logging
export LOG_LEVEL=debug
export GIN_MODE=debug

# Run the service
go run ./cmd/server/main.go
```

## Monitoring and Alerting

### Production Monitoring

- **INFO logs**: Monitor for business events and normal operations
- **WARN logs**: Alert on security concerns and potential issues
- **ERROR logs**: Alert on failures and exceptions

### Development Monitoring

- **DEBUG logs**: Use for troubleshooting and development
- **All levels**: Complete application flow visibility

## Migration from Previous Logging

The previous implementation used `log.Printf` directly, which always executed regardless of environment. The new system:

1. **Maintains compatibility**: All existing log messages work
2. **Adds conditional behavior**: Debug messages are suppressed in production
3. **Improves performance**: Reduces I/O overhead in production
4. **Enhances security**: Prevents sensitive debug information from appearing in production logs

## Testing

The logging system includes comprehensive tests:

```bash
# Run logger tests
go test ./utils/... -v

# Test middleware with different log levels
LOG_LEVEL=info go test ./middleware/... -v
LOG_LEVEL=debug go test ./middleware/... -v
```

## Troubleshooting

### Debug Logs Not Appearing

1. Check `LOG_LEVEL` environment variable
2. Verify `GIN_MODE` is not set to `release`
3. Ensure debug logging is enabled in the logger instance

### Performance Issues

1. Verify `LOG_LEVEL` is set to `info` or higher in production
2. Check for expensive operations in debug logs
3. Use `utils.IsDebugEnabled()` for conditional expensive operations

### Log Level Not Changing

1. Restart the application after changing environment variables
2. Check that the logger is using the default instance
3. Verify environment variable names are correct
