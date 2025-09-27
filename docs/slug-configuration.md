# Slug Configuration

This document describes the configurable slug suggestion system in the rtr-user-auth-service, which allows customization of tenant slug suggestions per deployment environment.

## Overview

The slug suggestion system generates alternative tenant slugs when the primary slug is already taken. Previously, this used hardcoded suffixes, but now it supports environment-based configuration for better customization across different deployment environments.

## Configuration

### Environment Variable

**`SLUG_SUGGESTION_SUFFIXES`**: Comma-separated list of suffixes to use for slug suggestions.

### Format

```bash
# Comma-separated suffixes (with or without leading dashes)
SLUG_SUGGESTION_SUFFIXES="-corp,-inc,-ltd,-app"

# Suffixes without dashes (automatically prefixed)
SLUG_SUGGESTION_SUFFIXES="corp,inc,ltd,app"

# Mixed format (some with dashes, some without)
SLUG_SUGGESTION_SUFFIXES="-corp,inc,-ltd,app"
```

### Default Behavior

If `SLUG_SUGGESTION_SUFFIXES` is not set or is empty, the system uses these default suffixes:

```go
defaultSlugSuffixes = []string{"-hq", "-io", "-team", "-app", "-co"}
```

## Examples

### Default Configuration

```bash
# No environment variable set
# Uses: ["-hq", "-io", "-team", "-app", "-co"]

# Input: "acme"
# Suggestions: ["acme-hq", "acme-io", "acme-team", "acme-app", "acme-co"]
```

### Custom Configuration

```bash
# Set custom suffixes
export SLUG_SUGGESTION_SUFFIXES="-corp,-inc,-ltd,-app"

# Input: "acme"
# Suggestions: ["acme-corp", "acme-inc", "acme-ltd", "acme-app"]
```

### Enterprise Configuration

```bash
# Enterprise-focused suffixes
export SLUG_SUGGESTION_SUFFIXES="-enterprise,-corp,-inc,-ltd,-group"

# Input: "acme"
# Suggestions: ["acme-enterprise", "acme-corp", "acme-inc", "acme-ltd", "acme-group"]
```

### Startup/Technology Configuration

```bash
# Tech startup focused
export SLUG_SUGGESTION_SUFFIXES="-tech,-labs,-studio,-works,-dev"

# Input: "acme"
# Suggestions: ["acme-tech", "acme-labs", "acme-studio", "acme-works", "acme-dev"]
```

## Deployment Examples

### Docker

```dockerfile
# Set custom slug suffixes for production
ENV SLUG_SUGGESTION_SUFFIXES="-corp,-inc,-ltd,-app"

# Or for development
ENV SLUG_SUGGESTION_SUFFIXES="-dev,-test,-local,-demo"
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
        - name: SLUG_SUGGESTION_SUFFIXES
          value: "-corp,-inc,-ltd,-app"
```

### Local Development

```bash
# Development environment
export SLUG_SUGGESTION_SUFFIXES="-dev,-test,-local"

# Production environment
export SLUG_SUGGESTION_SUFFIXES="-corp,-inc,-ltd"

# Run the service
go run ./cmd/server/main.go
```

## Implementation Details

### Function Behavior

```go
func SuggestSlugAlternatives(base string) []string {
    suffixes := getSlugSuffixes()  // Gets configured or default suffixes
    suggestions := make([]string, 0, len(suffixes))
    
    for _, suffix := range suffixes {
        candidate := base + suffix
        if len(candidate) <= 30 {  // Respects 30-character limit
            suggestions = append(suggestions, candidate)
        }
    }
    
    // Fallback to base slug if no suggestions generated
    if len(suggestions) == 0 {
        suggestions = append(suggestions, base)
    }
    
    return suggestions
}
```

### Environment Variable Parsing

The system automatically:

1. **Trims whitespace** from each suffix
2. **Adds leading dash** if not present
3. **Filters empty values** from the list
4. **Falls back to defaults** if no valid suffixes found

### Validation

- **Length limit**: All suggestions must be ≤ 30 characters
- **Format validation**: Suffixes are normalized to start with a dash
- **Fallback behavior**: Returns base slug if no valid suggestions can be generated

## API Usage

### Tenant Creation Response

When a slug conflict occurs during tenant creation, the API returns suggestions:

```json
{
  "code": "TENANT_SLUG_TAKEN",
  "message": "Tenant slug 'acme' is already taken",
  "suggestions": [
    "acme-corp",
    "acme-inc", 
    "acme-ltd",
    "acme-app"
  ]
}
```

### Error Handling

```go
// In tenant service
suggestions := utils.SuggestSlugAlternatives(baseSlug)
return &domain.ErrTenantSlugTaken{
    Slug: baseSlug,
    Suggestions: suggestions,
}
```

## Testing

### Unit Tests

The system includes comprehensive tests for:

- **Default behavior**: Uses default suffixes when no environment variable is set
- **Custom configuration**: Parses environment variable correctly
- **Format handling**: Handles various input formats (with/without dashes, spaces, etc.)
- **Edge cases**: Empty values, invalid input, fallback behavior
- **Length validation**: Ensures suggestions don't exceed character limits

### Running Tests

```bash
# Run slug-related tests
go test ./utils/... -v -run="TestSuggestSlugAlternatives"

# Run all slug tests
go test ./utils/... -v -run="TestNormalizeSlug|TestSuggestSlugAlternatives"
```

## Best Practices

### 1. Choose Relevant Suffixes

```bash
# ✅ Good: Business-relevant suffixes
SLUG_SUGGESTION_SUFFIXES="-corp,-inc,-ltd,-group"

# ❌ Avoid: Generic or confusing suffixes
SLUG_SUGGESTION_SUFFIXES="-x,-y,-z,-temp"
```

### 2. Consider Your Domain

```bash
# Enterprise/B2B focus
SLUG_SUGGESTION_SUFFIXES="-enterprise,-corp,-inc,-ltd"

# Technology/Startup focus  
SLUG_SUGGESTION_SUFFIXES="-tech,-labs,-studio,-works"

# Regional focus
SLUG_SUGGESTION_SUFFIXES="-us,-eu,-asia,-global"
```

### 3. Keep Suffixes Short

```bash
# ✅ Good: Short, clear suffixes
SLUG_SUGGESTION_SUFFIXES="-corp,-inc,-ltd"

# ❌ Avoid: Long suffixes that reduce available space
SLUG_SUGGESTION_SUFFIXES="-corporation,-incorporated,-limited"
```

### 4. Test Your Configuration

```bash
# Test your configuration
export SLUG_SUGGESTION_SUFFIXES="-corp,-inc,-ltd"
go test ./utils/... -v -run="TestSuggestSlugAlternatives"
```

## Migration Guide

### From Hardcoded to Configurable

**Before** (hardcoded):
```go
suffixes := []string{"-hq", "-io", "-team"}
```

**After** (configurable):
```go
suffixes := getSlugSuffixes()  // Reads from environment or uses defaults
```

### Deployment Updates

1. **Set environment variable** in your deployment configuration
2. **Test the configuration** with your specific suffixes
3. **Update documentation** to reflect your custom suffixes
4. **Monitor tenant creation** to ensure suggestions are appropriate

## Troubleshooting

### Common Issues

1. **No suggestions generated**: Check that suffixes don't make slugs exceed 30 characters
2. **Wrong suffixes appearing**: Verify `SLUG_SUGGESTION_SUFFIXES` environment variable is set correctly
3. **Fallback to defaults**: Ensure environment variable is not empty or contains only invalid values

### Debugging

```bash
# Check current configuration
echo $SLUG_SUGGESTION_SUFFIXES

# Test with specific input
go test ./utils/... -v -run="TestSuggestSlugAlternatives"
```

### Validation

The system automatically validates:
- Suffix format (adds dashes if missing)
- Length constraints (30 character limit)
- Empty value filtering
- Fallback to defaults when needed
