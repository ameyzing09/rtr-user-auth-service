package middleware

import (
	"net/http"
	"os"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/services"
	"rtr-user-auth-service/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		utils.Debug("[AuthMiddleware] Authorization header present=%t", authHeader != "")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.Debug("[AuthMiddleware] Missing or invalid Authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		utils.Debug("[AuthMiddleware] Received token length=%d", len(tokenStr))

		claims := &utils.Claims{}
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			utils.Warn("[AuthMiddleware] JWT_SECRET not set, using default (development) secret")
			secret = "default_secret"
		}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			utils.Debug("[AuthMiddleware] Token validation failed: error=%v, valid=%t", err, token != nil && token.Valid)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		actor := services.UserRead{
			ID:       claims.UserID,
			TenantID: claims.TenantID,
			Email:    claims.Email,
			Role:     models.Role(claims.Role),
		}

		utils.Debug("[AuthMiddleware] Authenticated actor: userID=%s tenantID=%s role=%s", actor.ID, actor.TenantID, actor.Role)

		// Check tenant boundary enforcement
		if tid := c.GetString(CtxTenantIDKey); tid != "" && actor.Role != models.RoleSuperAdmin && tid != actor.TenantID {
			utils.Warn("[AuthMiddleware] Tenant mismatch: requestTenant=%s actorTenant=%s actorRole=%s", tid, actor.TenantID, actor.Role)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access to this tenant is forbidden"})
			return
		}

		c.Set("actor", actor)
		utils.Debug("[AuthMiddleware] Actor set in context, proceeding to next handler")
		c.Next()
	}
}
