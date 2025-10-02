package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"rtr-user-auth-service/models"
	"rtr-user-auth-service/services"
	"rtr-user-auth-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const controlPlaneContextKey = "__control_plane__"

func ControlPlaneScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		//print log
		utils.Debug("[ControlPlaneScope] Marking request as control-plane scope")
		c.Set(controlPlaneContextKey, true)
		c.Next()
	}
}

func isControlPlaneRequest(c *gin.Context) bool {
	//print log
	utils.Debug("[isControlPlaneRequest] Checking if request is control-plane scope")
	value, ok := c.Get(controlPlaneContextKey)
	if !ok {
		return false
	}
	flag, _ := value.(bool)
	return flag
}

func permitDevSuperadmin(c *gin.Context, token string) bool {
	if !isControlPlaneRequest(c) {
		return false
	}
	//print log
	utils.Debug("[AuthMiddleware] Control-plane request detected, checking for SUPERADMIN_DEV_TOKEN")

	env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	//print env
	utils.Debug("[AuthMiddleware] Environment: %s", env)
	if env == "" {
		env = "local"
	}
	if env != "local" && env != "dev" {
		return false
	}

	fallback := strings.TrimSpace(os.Getenv("SUPERADMIN_DEV_TOKEN"))
	if fallback == "" {
		fallback = "dev-superadmin"
	}

	if subtle.ConstantTimeCompare([]byte(token), []byte(fallback)) != 1 {
		return false
	}

	utils.Warn("[AuthMiddleware] SUPERADMIN_DEV_TOKEN accepted for control-plane request (env=%s)", env)

	actor := services.UserRead{
		ID:       "dev-superadmin",
		TenantID: "",
		Email:    "dev-superadmin@local",
		Role:     models.RoleSuperAdmin,
	}

	c.Set("actor", actor)
	c.Next()
	return true
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		utils.Debug("[AuthMiddleware] Authorization header present=%t", authHeader != "")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.Debug("[AuthMiddleware] Missing or invalid Authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			return
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
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

		utils.Debug("[AuthMiddleware] Parsed token: token=%v, err=%v", token, err)
		if err != nil || !token.Valid {
			if permitDevSuperadmin(c, tokenStr) {
				return
			}

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
