package middleware

import (
	"fmt"
	"net/http"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/services"
	"rtr-user-auth-service/utils"

	"github.com/gin-gonic/gin"
)

// RequirePermission creates a middleware that requires a specific permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		actorValue, exists := c.Get("actor")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		actor, ok := actorValue.(services.UserRead)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid actor context"})
			return
		}

		if !models.HasPermission(actor.Permissions, permission) {
			// Audit: Permission denied
			if auditSvc := GetAuditService(c); auditSvc != nil {
				clientIP := GetClientIP(c)
				userAgent := GetUserAgent(c)
				reason := fmt.Sprintf("Missing permission: %s", permission)
				actorRoleStr := string(actor.Role)
				_ = auditSvc.Log(c.Request.Context(), services.AuditLogEntry{
					Action:        utils.AuditActionPermissionDenied,
					ActorID:       &actor.ID,
					ActorTenantID: &actor.TenantID,
					ActorRole:     &actorRoleStr,
					Status:        models.AuditStatusDenied,
					Reason:        &reason,
					IPAddress:     utils.StringPtr(clientIP),
					UserAgent:     utils.StringPtr(userAgent),
					Metadata: map[string]interface{}{
						"required_permission": permission,
						"path":                c.Request.URL.Path,
						"method":              c.Request.Method,
					},
				})
			}

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden", "required": permission})
			return
		}

		c.Next()
	}
}

// RequireAnyPermission creates a middleware that requires at least one of the specified permissions
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		actorValue, exists := c.Get("actor")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		actor, ok := actorValue.(services.UserRead)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid actor context"})
			return
		}

		if !models.HasAnyPermission(actor.Permissions, permissions...) {
			// Audit: Permission denied
			if auditSvc := GetAuditService(c); auditSvc != nil {
				clientIP := GetClientIP(c)
				userAgent := GetUserAgent(c)
				reason := fmt.Sprintf("Missing any of permissions: %v", permissions)
				actorRoleStr := string(actor.Role)
				_ = auditSvc.Log(c.Request.Context(), services.AuditLogEntry{
					Action:        utils.AuditActionPermissionDenied,
					ActorID:       &actor.ID,
					ActorTenantID: &actor.TenantID,
					ActorRole:     &actorRoleStr,
					Status:        models.AuditStatusDenied,
					Reason:        &reason,
					IPAddress:     utils.StringPtr(clientIP),
					UserAgent:     utils.StringPtr(userAgent),
					Metadata: map[string]interface{}{
						"required_any_permission": permissions,
						"path":                    c.Request.URL.Path,
						"method":                  c.Request.Method,
					},
				})
			}

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden", "required_any": permissions})
			return
		}

		c.Next()
	}
}

// RequireAllPermissions creates a middleware that requires all of the specified permissions
func RequireAllPermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		actorValue, exists := c.Get("actor")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		actor, ok := actorValue.(services.UserRead)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid actor context"})
			return
		}

		if !models.HasAllPermissions(actor.Permissions, permissions...) {
			// Audit: Permission denied
			if auditSvc := GetAuditService(c); auditSvc != nil {
				clientIP := GetClientIP(c)
				userAgent := GetUserAgent(c)
				reason := fmt.Sprintf("Missing all permissions: %v", permissions)
				actorRoleStr := string(actor.Role)
				_ = auditSvc.Log(c.Request.Context(), services.AuditLogEntry{
					Action:        utils.AuditActionPermissionDenied,
					ActorID:       &actor.ID,
					ActorTenantID: &actor.TenantID,
					ActorRole:     &actorRoleStr,
					Status:        models.AuditStatusDenied,
					Reason:        &reason,
					IPAddress:     utils.StringPtr(clientIP),
					UserAgent:     utils.StringPtr(userAgent),
					Metadata: map[string]interface{}{
						"required_all_permissions": permissions,
						"path":                     c.Request.URL.Path,
						"method":                   c.Request.Method,
					},
				})
			}

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden", "required_all": permissions})
			return
		}

		c.Next()
	}
}
