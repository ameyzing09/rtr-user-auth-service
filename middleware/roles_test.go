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

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		actorRole      models.Role
		requiredRole   models.Role
		expectedStatus int
	}{
		{
			name:           "SUPERADMIN can access SUPERADMIN route",
			actorRole:      models.RoleSuperAdmin,
			requiredRole:   models.RoleSuperAdmin,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ADMIN cannot access SUPERADMIN route",
			actorRole:      models.RoleAdmin,
			requiredRole:   models.RoleSuperAdmin,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "ADMIN can access ADMIN route",
			actorRole:      models.RoleAdmin,
			requiredRole:   models.RoleAdmin,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "HR cannot access ADMIN route",
			actorRole:      models.RoleHR,
			requiredRole:   models.RoleAdmin,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				// Set actor in context
				actor := services.UserRead{
					ID:       "test-user",
					TenantID: "test-tenant",
					Email:    "test@example.com",
					Role:     tt.actorRole,
				}
				c.Set("actor", actor)
				c.Next()
			})
			router.GET("/test", RequireRole(tt.requiredRole), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRequireAny(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		actorRole      models.Role
		allowedRoles   []models.Role
		expectedStatus int
	}{
		{
			name:           "ADMIN can access ADMIN or HR route",
			actorRole:      models.RoleAdmin,
			allowedRoles:   []models.Role{models.RoleAdmin, models.RoleHR},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "HR can access ADMIN or HR route",
			actorRole:      models.RoleHR,
			allowedRoles:   []models.Role{models.RoleAdmin, models.RoleHR},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "INTERVIEWER cannot access ADMIN or HR route",
			actorRole:      models.RoleInterviewer,
			allowedRoles:   []models.Role{models.RoleAdmin, models.RoleHR},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "CANDIDATE cannot access ADMIN or HR route",
			actorRole:      models.RoleCandidate,
			allowedRoles:   []models.Role{models.RoleAdmin, models.RoleHR},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				// Set actor in context
				actor := services.UserRead{
					ID:       "test-user",
					TenantID: "test-tenant",
					Email:    "test@example.com",
					Role:     tt.actorRole,
				}
				c.Set("actor", actor)
				c.Next()
			})
			router.GET("/test", RequireAny(tt.allowedRoles...), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRequireRole_NoActor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", RequireRole(models.RoleAdmin), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
