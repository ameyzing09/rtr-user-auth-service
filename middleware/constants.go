package middleware

// HTTP Header names
const (
	HeaderAuthorization  = "Authorization"
	HeaderTenantID       = "X-Tenant-ID"
	HeaderTenantSlug     = "X-Tenant-Slug"
	HeaderIdempotencyKey = "Idempotency-Key"
	HeaderContentType    = "Content-Type"
	HeaderCacheControl   = "Cache-Control"
	HeaderPragma         = "Pragma"
)

// Context keys
const (
	ContextKeyActor  = "actor"
	ContextKeyTenant = "tenant"
)

// Authorization schemes
const (
	BearerScheme = "Bearer "
	DevToken     = "dev-token"
)

// Cache control values
const (
	CacheControlNoStore = "no-store"
	PragmaNoCache       = "no-cache"
)
