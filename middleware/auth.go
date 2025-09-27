package middleware

import (
	"log"
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
		log.Printf("[AuthMiddleware] Authorization header present=%t", authHeader != "")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			log.Printf("[AuthMiddleware] Missing or invalid Authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		log.Printf("[AuthMiddleware] Received token length=%d", len(tokenStr))
		claims := &utils.Claims{}
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			log.Printf("[AuthMiddleware] JWT_SECRET not set, using default (development) secret")
			secret = "default_secret"
		}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		// print token validity and non-sensitive claims
		if token != nil && token.Valid {
			log.Printf("[AuthMiddleware] Token valid: true, userID=%s tenantID=%s role=%s", claims.UserID, claims.TenantID, claims.Role)
		} else {
			log.Printf("[AuthMiddleware] Token valid: false")
		}
		log.Printf("[AuthMiddleware] Parse error: %v", err)

		if err != nil || !token.Valid {
			log.Printf("[AuthMiddleware] token parse error=%v, valid=%t", err, token != nil && token.Valid)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		actor := services.UserRead{
			ID:       claims.UserID,
			TenantID: claims.TenantID,
			Email:    claims.Email,
			Role:     models.Role(claims.Role),
		}
		log.Printf("[AuthMiddleware] Authenticated actor: userID=%s tenantID=%s role=%s", actor.ID, actor.TenantID, actor.Role)
		if tid := c.GetString(CtxTenantIDKey); tid != "" && actor.Role != models.RoleSuperAdmin && tid != actor.TenantID {
			log.Printf("[AuthMiddleware] Tenant mismatch: requestTenant=%s actorTenant=%s actorRole=%s", tid, actor.TenantID, actor.Role)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access to this tenant is forbidden"})
			return
		}
		c.Set("actor", actor)
		log.Printf("[AuthMiddleware] Actor set in context, proceeding to next handler")
		c.Next()
	}
}
