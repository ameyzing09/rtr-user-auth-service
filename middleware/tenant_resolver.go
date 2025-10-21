package middleware

import "github.com/gin-gonic/gin"

const (
	CtxTenantIDKey = "tenantID"
	CtxTenantKey   = "tenant"
)

// GetTenantIDFromContext extracts the tenant ID from the Gin context
// Returns empty string if tenant ID is not found in context
func GetTenantIDFromContext(c *gin.Context) string {
	return c.GetString(CtxTenantIDKey)
}
