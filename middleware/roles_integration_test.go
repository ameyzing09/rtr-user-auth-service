package middleware

import (
	"net/http"
	"net/http/httptest"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/services"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRoleGates_WithTenantContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		userRole        models.Role
		userTenantID    string
		contextTenantID string
		requiredRoles   []models.Role
		expectedStatus  int
		description     string
	}{
		// RequireRole tests
		{
			name:            "ADMIN can access ADMIN-only route in same tenant",
			userRole:        models.RoleAdmin,
			userTenantID:    "tenant-a",
			contextTenantID: "tenant-a",
			requiredRoles:   []models.Role{models.RoleAdmin},
			expectedStatus:  http.StatusOK,
			description:     "ADMIN can access their own tenant with correct role",
		},
		{
			name:            "HR cannot access ADMIN-only route",
			userRole:        models.RoleHR,
			userTenantID:    "tenant-a",
			contextTenantID: "tenant-a",
			requiredRoles:   []models.Role{models.RoleAdmin},
			expectedStatus:  http.StatusForbidden,
			description:     "HR cannot access ADMIN-only routes",
		},
		{
			name:            "ADMIN cannot access route in different tenant",
			userRole:        models.RoleAdmin,
			userTenantID:    "tenant-a",
			contextTenantID: "tenant-b",
			requiredRoles:   []models.Role{models.RoleAdmin},
			expectedStatus:  http.StatusForbidden,
			description:     "ADMIN cannot cross tenant boundaries",
		},
		{
			name:            "SUPERADMIN can access ADMIN route in any tenant",
			userRole:        models.RoleSuperAdmin,
			userTenantID:    "tenant-a",
			contextTenantID: "tenant-b",
			requiredRoles:   []models.Role{models.RoleAdmin},
			expectedStatus:  http.StatusOK,
			description:     "SUPERADMIN can bypass tenant boundaries",
		},

		// RequireAny tests
		{
			name:            "ADMIN can access ADMIN or HR route",
			userRole:        models.RoleAdmin,
			userTenantID:    "tenant-a",
			contextTenantID: "tenant-a",
			requiredRoles:   []models.Role{models.RoleAdmin, models.RoleHR},
			expectedStatus:  http.StatusOK,
			description:     "ADMIN can access routes requiring ADMIN or HR",
		},
		{
			name:            "HR can access ADMIN or HR route",
			userRole:        models.RoleHR,
			userTenantID:    "tenant-a",
			contextTenantID: "tenant-a",
			requiredRoles:   []models.Role{models.RoleAdmin, models.RoleHR},
			expectedStatus:  http.StatusOK,
			description:     "HR can access routes requiring ADMIN or HR",
		},
		{
			name:            "INTERVIEWER cannot access ADMIN or HR route",
			userRole:        models.RoleInterviewer,
			userTenantID:    "tenant-a",
			contextTenantID: "tenant-a",
			requiredRoles:   []models.Role{models.RoleAdmin, models.RoleHR},
			expectedStatus:  http.StatusForbidden,
			description:     "INTERVIEWER cannot access ADMIN or HR routes",
		},
		{
			name:            "CANDIDATE cannot access ADMIN or HR route",
			userRole:        models.RoleCandidate,
			userTenantID:    "tenant-a",
			contextTenantID: "tenant-a",
			requiredRoles:   []models.Role{models.RoleAdmin, models.RoleHR},
			expectedStatus:  http.StatusForbidden,
			description:     "CANDIDATE cannot access ADMIN or HR routes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router with tenant context and auth middleware
			router := gin.New()
			router.Use(func(c *gin.Context) {
				// Simulate tenant context being set
				c.Set(CtxTenantIDKey, tt.contextTenantID)
				c.Next()
			})
			router.Use(func(c *gin.Context) {
				// Set actor in context (simulating auth middleware)
				actor := services.UserRead{
					ID:       "user-123",
					TenantID: tt.userTenantID,
					Email:    "test@example.com",
					Role:     tt.userRole,
				}
				c.Set("actor", actor)

				// Simulate tenant boundary check from auth middleware
				// SUPERADMIN can bypass tenant boundaries, others cannot
				if tid := c.GetString(CtxTenantIDKey); tid != "" && actor.Role != models.RoleSuperAdmin && tid != actor.TenantID {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access to this tenant is forbidden"})
					return
				}

				c.Next()
			})

			// Apply role gate based on test
			if len(tt.requiredRoles) == 1 {
				router.GET("/test", RequireRole(tt.requiredRoles[0]), func(c *gin.Context) {
					c.Status(http.StatusOK)
				})
			} else {
				router.GET("/test", RequireAny(tt.requiredRoles...), func(c *gin.Context) {
					c.Status(http.StatusOK)
				})
			}

			// Make request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)
		})
	}
}

func TestRoleGates_NoActor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requiredRoles  []models.Role
		expectedStatus int
		description    string
	}{
		{
			name:           "RequireRole with no actor returns 401",
			requiredRoles:  []models.Role{models.RoleAdmin},
			expectedStatus: http.StatusUnauthorized,
			description:    "No actor context should return 401",
		},
		{
			name:           "RequireAny with no actor returns 401",
			requiredRoles:  []models.Role{models.RoleAdmin, models.RoleHR},
			expectedStatus: http.StatusUnauthorized,
			description:    "No actor context should return 401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router without actor context
			router := gin.New()

			// Apply role gate based on test
			if len(tt.requiredRoles) == 1 {
				router.GET("/test", RequireRole(tt.requiredRoles[0]), func(c *gin.Context) {
					c.Status(http.StatusOK)
				})
			} else {
				router.GET("/test", RequireAny(tt.requiredRoles...), func(c *gin.Context) {
					c.Status(http.StatusOK)
				})
			}

			// Make request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)
		})
	}
}

func TestRoleGates_InvalidActorType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup router with invalid actor type
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Set invalid actor type
		c.Set("actor", "invalid-type")
		c.Next()
	})
	router.GET("/test", RequireRole(models.RoleAdmin), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Invalid actor type should return 401")
}

func TestRoleGates_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userRole       models.Role
		requiredRoles  []models.Role
		expectedStatus int
		description    string
	}{
		{
			name:           "SUPERADMIN can access any role requirement",
			userRole:       models.RoleSuperAdmin,
			requiredRoles:  []models.Role{models.RoleCandidate},
			expectedStatus: http.StatusOK,
			description:    "SUPERADMIN should be able to access any role-gated route",
		},
		{
			name:           "Empty RequireAny should deny all",
			userRole:       models.RoleAdmin,
			requiredRoles:  []models.Role{},
			expectedStatus: http.StatusForbidden,
			description:    "Empty RequireAny should deny all roles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router with actor context
			router := gin.New()
			router.Use(func(c *gin.Context) {
				actor := services.UserRead{
					ID:       "user-123",
					TenantID: "tenant-a",
					Email:    "test@example.com",
					Role:     tt.userRole,
				}
				c.Set("actor", actor)
				c.Next()
			})

			// Apply role gate
			if len(tt.requiredRoles) == 1 {
				router.GET("/test", RequireRole(tt.requiredRoles[0]), func(c *gin.Context) {
					c.Status(http.StatusOK)
				})
			} else {
				router.GET("/test", RequireAny(tt.requiredRoles...), func(c *gin.Context) {
					c.Status(http.StatusOK)
				})
			}

			// Make request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)
		})
	}
}
