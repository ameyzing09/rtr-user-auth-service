package middleware

import (
	"net/http"
	"net/http/httptest"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/services"
	"rtr-user-auth-service/utils"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// Mock JWT secret for testing
const testJWTSecret = "test_secret_key"

// createTestToken creates a JWT token for testing
func createTestToken(userID, tenantID, email string, role models.Role) string {
	claims := &utils.Claims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    email,
		Role:     string(role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testJWTSecret))
	return tokenString
}

func TestAuthMiddleware_TenantMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set JWT secret for testing
	t.Setenv("JWT_SECRET", testJWTSecret)

	tests := []struct {
		name            string
		tokenTenantID   string
		contextTenantID string
		role            models.Role
		expectedStatus  int
		description     string
	}{
		{
			name:            "ADMIN from tenant A cannot access tenant B",
			tokenTenantID:   "tenant-a",
			contextTenantID: "tenant-b",
			role:            models.RoleAdmin,
			expectedStatus:  http.StatusForbidden,
			description:     "Regular users cannot cross tenant boundaries",
		},
		{
			name:            "HR from tenant A cannot access tenant B",
			tokenTenantID:   "tenant-a",
			contextTenantID: "tenant-b",
			role:            models.RoleHR,
			expectedStatus:  http.StatusForbidden,
			description:     "HR users cannot cross tenant boundaries",
		},
		{
			name:            "INTERVIEWER from tenant A cannot access tenant B",
			tokenTenantID:   "tenant-a",
			contextTenantID: "tenant-b",
			role:            models.RoleInterviewer,
			expectedStatus:  http.StatusForbidden,
			description:     "Interviewers cannot cross tenant boundaries",
		},
		{
			name:            "CANDIDATE from tenant A cannot access tenant B",
			tokenTenantID:   "tenant-a",
			contextTenantID: "tenant-b",
			role:            models.RoleCandidate,
			expectedStatus:  http.StatusForbidden,
			description:     "Candidates cannot cross tenant boundaries",
		},
		{
			name:            "SUPERADMIN can access any tenant",
			tokenTenantID:   "tenant-a",
			contextTenantID: "tenant-b",
			role:            models.RoleSuperAdmin,
			expectedStatus:  http.StatusOK,
			description:     "SUPERADMIN can bypass tenant boundaries",
		},
		{
			name:            "ADMIN can access their own tenant",
			tokenTenantID:   "tenant-a",
			contextTenantID: "tenant-a",
			role:            models.RoleAdmin,
			expectedStatus:  http.StatusOK,
			description:     "Users can access their own tenant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test token
			token := createTestToken("user-123", tt.tokenTenantID, "test@example.com", tt.role)

			// Setup router with auth middleware
			router := gin.New()
			router.Use(func(c *gin.Context) {
				// Simulate tenant context being set
				c.Set(CtxTenantIDKey, tt.contextTenantID)
				c.Next()
			})
			router.Use(AuthMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// Make request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)
		})
	}
}

func TestAuthMiddleware_NoTenantContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set JWT secret for testing
	t.Setenv("JWT_SECRET", testJWTSecret)

	tests := []struct {
		name           string
		role           models.Role
		expectedStatus int
		description    string
	}{
		{
			name:           "ADMIN without tenant context should succeed",
			role:           models.RoleAdmin,
			expectedStatus: http.StatusOK,
			description:    "No tenant context means no tenant boundary check",
		},
		{
			name:           "SUPERADMIN without tenant context should succeed",
			role:           models.RoleSuperAdmin,
			expectedStatus: http.StatusOK,
			description:    "SUPERADMIN can access routes without tenant context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test token
			token := createTestToken("user-123", "tenant-a", "test@example.com", tt.role)

			// Setup router with auth middleware (no tenant context)
			router := gin.New()
			router.Use(AuthMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// Make request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set JWT secret for testing
	t.Setenv("JWT_SECRET", testJWTSecret)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		description    string
	}{
		{
			name:           "Missing Authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			description:    "Missing auth header should return 401",
		},
		{
			name:           "Invalid Bearer format",
			authHeader:     "Invalid token",
			expectedStatus: http.StatusUnauthorized,
			description:    "Invalid bearer format should return 401",
		},
		{
			name:           "Invalid JWT token",
			authHeader:     "Bearer invalid.jwt.token",
			expectedStatus: http.StatusUnauthorized,
			description:    "Invalid JWT should return 401",
		},
		{
			name:           "Expired token",
			authHeader:     "Bearer " + createExpiredToken(),
			expectedStatus: http.StatusUnauthorized,
			description:    "Expired token should return 401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router with auth middleware
			router := gin.New()
			router.Use(AuthMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// Make request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)
		})
	}
}

func TestAuthMiddleware_ActorContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set JWT secret for testing
	t.Setenv("JWT_SECRET", testJWTSecret)

	// Create test token
	token := createTestToken("user-123", "tenant-a", "test@example.com", models.RoleAdmin)

	// Setup router with auth middleware
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		// Check if actor is set in context
		actorValue, exists := c.Get("actor")
		assert.True(t, exists, "Actor should be set in context")

		actor, ok := actorValue.(services.UserRead)
		assert.True(t, ok, "Actor should be of type services.UserRead")
		assert.Equal(t, "user-123", actor.ID)
		assert.Equal(t, "tenant-a", actor.TenantID)
		assert.Equal(t, "test@example.com", actor.Email)
		assert.Equal(t, models.RoleAdmin, actor.Role)

		c.Status(http.StatusOK)
	})

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
}

// Helper function to create an expired token
func createExpiredToken() string {
	claims := &utils.Claims{
		UserID:   "user-123",
		TenantID: "tenant-a",
		Email:    "test@example.com",
		Role:     string(models.RoleAdmin),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testJWTSecret))
	return tokenString
}
