package middleware

import (
	"net/http"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/services"

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
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden", "required_all": permissions})
			return
		}

		c.Next()
	}
}
