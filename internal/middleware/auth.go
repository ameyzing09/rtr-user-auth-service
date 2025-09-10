package middleware

import (
	"net/http"
	"strings"

	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/entities"
	"github.com/ameyzing09/rtr-user-auth-service/internal/services"
	"github.com/ameyzing09/rtr-user-auth-service/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	UserContextKey     = "user"
	TenantContextKey   = "tenant"
	ClaimsContextKey   = "claims"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Token is required",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set claims in context
		c.Set(ClaimsContextKey, claims)
		c.Set(UserContextKey, claims.UserID)
		c.Set(TenantContextKey, claims.TenantID)

		c.Next()
	}
}

// RequireRole creates a middleware that requires specific roles
func RequireRole(roles ...entities.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get(ClaimsContextKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*utils.JWTClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Invalid claims type",
			})
			c.Abort()
			return
		}

		// Check if user has one of the required roles
		hasRole := false
		for _, role := range roles {
			if userClaims.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin creates a middleware that requires admin role
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(entities.RoleAdmin)
}

// RequireAdminOrHR creates a middleware that requires admin or HR role
func RequireAdminOrHR() gin.HandlerFunc {
	return RequireRole(entities.RoleAdmin, entities.RoleHR)
}

// RequireSameTenantOrAdmin creates a middleware that requires same tenant or admin role
func RequireSameTenantOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get(ClaimsContextKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*utils.JWTClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Invalid claims type",
			})
			c.Abort()
			return
		}

		// Admin can access any tenant
		if userClaims.Role == entities.RoleAdmin {
			c.Next()
			return
		}

		// Get tenant ID from URL parameter
		tenantIDParam := c.Param("tenantId")
		if tenantIDParam == "" {
			// If no tenant ID in URL, check if user is accessing their own tenant
			c.Next()
			return
		}

		requestedTenantID, err := uuid.Parse(tenantIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "Invalid tenant ID format",
			})
			c.Abort()
			return
		}

		// Check if user belongs to the requested tenant
		if userClaims.TenantID != requestedTenantID {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Access denied to this tenant",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSameUserOrAdminOrHR creates a middleware that requires same user, admin, or HR role
func RequireSameUserOrAdminOrHR() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get(ClaimsContextKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*utils.JWTClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Invalid claims type",
			})
			c.Abort()
			return
		}

		// Admin and HR can access any user
		if userClaims.Role == entities.RoleAdmin || userClaims.Role == entities.RoleHR {
			c.Next()
			return
		}

		// Get user ID from URL parameter
		userIDParam := c.Param("userId")
		if userIDParam == "" {
			// If no user ID in URL, allow access (user accessing their own profile)
			c.Next()
			return
		}

		requestedUserID, err := uuid.Parse(userIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "Invalid user ID format",
			})
			c.Abort()
			return
		}

		// Check if user is accessing their own profile
		if userClaims.UserID != requestedUserID {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Access denied to this user",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}