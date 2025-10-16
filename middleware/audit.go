package middleware

import (
	"rtr-user-auth-service/services"
	"rtr-user-auth-service/utils"

	"github.com/gin-gonic/gin"
)

const (
	// Context keys for audit logging
	CtxAuditServiceKey = "audit_service"
	CtxClientIPKey     = "client_ip"
	CtxUserAgentKey    = "user_agent"
)

// AuditMiddleware extracts client IP and User-Agent and stores them in context
// This should be applied early in the middleware chain, before auth
func AuditMiddleware(auditSvc services.AuditLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store audit service in context
		c.Set(CtxAuditServiceKey, auditSvc)

		// Extract and store client metadata
		clientIP := utils.ExtractClientIP(c.Request)
		userAgent := utils.ExtractUserAgent(c.Request)

		c.Set(CtxClientIPKey, clientIP)
		c.Set(CtxUserAgentKey, userAgent)

		c.Next()
	}
}

// GetAuditService retrieves the audit service from the context
func GetAuditService(c *gin.Context) services.AuditLogService {
	if svc, exists := c.Get(CtxAuditServiceKey); exists {
		if auditSvc, ok := svc.(services.AuditLogService); ok {
			return auditSvc
		}
	}
	return nil
}

// GetClientIP retrieves the client IP from the context
func GetClientIP(c *gin.Context) string {
	if ip, exists := c.Get(CtxClientIPKey); exists {
		if ipStr, ok := ip.(string); ok {
			return ipStr
		}
	}
	return ""
}

// GetUserAgent retrieves the user agent from the context
func GetUserAgent(c *gin.Context) string {
	if ua, exists := c.Get(CtxUserAgentKey); exists {
		if uaStr, ok := ua.(string); ok {
			return uaStr
		}
	}
	return ""
}
