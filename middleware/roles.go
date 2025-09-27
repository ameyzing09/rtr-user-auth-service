package middleware

import (
	"net/http"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/services"

	"github.com/gin-gonic/gin"
)

// RequireRole creates a middleware that requires a specific role
func RequireRole(role models.Role) gin.HandlerFunc {
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

		// SUPERADMIN can access any role-gated route
		if actor.Role == models.RoleSuperAdmin {
			c.Next()
			return
		}

		if actor.Role != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Next()
	}
}

// RequireAny creates a middleware that requires any of the specified roles
func RequireAny(roles ...models.Role) gin.HandlerFunc {
	allowed := make(map[models.Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

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

		// SUPERADMIN can access any role-gated route
		if actor.Role == models.RoleSuperAdmin {
			c.Next()
			return
		}

		if _, ok := allowed[actor.Role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Next()
	}
}
