package middleware

import (
	"net/http"
	"rtr-user-auth-service/utils"

	"github.com/gin-gonic/gin"
)

const (
	CSRFTokenCookieName = "csrf_token"
	CSRFTokenHeaderName = "X-CSRF-Token"
)

// CSRFProtection validates CSRF tokens for state-changing requests
// Uses double-submit cookie pattern:
// - Token stored in cookie (readable by JS)
// - Same token sent in X-CSRF-Token header
// - Backend validates they match
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF validation for safe methods (RFC 7231)
		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		// For state-changing methods (POST, PUT, DELETE, PATCH), validate CSRF token
		headerToken := c.GetHeader(CSRFTokenHeaderName)
		cookieToken, err := c.Cookie(CSRFTokenCookieName)

		// If cookie-based auth is not being used, skip CSRF validation
		// (backwards compatibility with Authorization header only)
		if err != nil {
			// No CSRF cookie present - check if using cookie-based auth
			if _, cookieErr := c.Cookie("access_token"); cookieErr != nil {
				// No access_token cookie either - using Authorization header, skip CSRF
				c.Next()
				return
			}
			// Has access_token cookie but no CSRF token - invalid
			utils.Warn("[CSRF] Missing CSRF token cookie for cookie-based auth")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "CSRF token required for cookie-based authentication",
			})
			return
		}

		// Validate CSRF token
		if !utils.ValidateCSRFToken(headerToken, cookieToken) {
			utils.Warn("[CSRF] CSRF token validation failed: header=%s cookie=%s",
				maskToken(headerToken), maskToken(cookieToken))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "CSRF token validation failed",
			})
			return
		}

		utils.Debug("[CSRF] CSRF token validated successfully")
		c.Next()
	}
}

// maskToken masks token for logging (shows first 8 chars only)
func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:8] + "***"
}
