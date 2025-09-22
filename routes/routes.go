package routes

import (
	"net/http"
	"rtr-user-auth-service/handlers"
	"rtr-user-auth-service/middleware"
	"rtr-user-auth-service/repositories"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, userHandler *handlers.UserHandler, tenantSettingHandler *handlers.TenantSettingHandler, tenantHandler *handlers.TenantHandler, tenantRepo repositories.TenantRepository) {
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "rtr-user-auth-service: ok")
	})

	publicRoute := r.Group("/")
	publicRoute.Use(middleware.TenantContext(tenantRepo))
	{
		publicRoute.POST("/login", userHandler.Login)

		tenantScope := publicRoute.Group("/tenant")
		tenantScope.GET("/settings", tenantSettingHandler.Get)
	}

	protectedRoute := r.Group("/")
	protectedRoute.Use(middleware.TenantContext(tenantRepo), middleware.AuthMiddleware())
	{
		protectedRoute.GET("/me", userHandler.GetMe)
		protectedRoute.POST("/me/change-password", userHandler.ChangePassword)

		protectedRoute.GET("/users", userHandler.ListUsers)
		protectedRoute.POST("/users", userHandler.CreateUser)

		tenantScope := protectedRoute.Group("/tenant")
		tenantScope.PUT("/settings", tenantSettingHandler.Put)
	}

	// SUPERADMIN-only API routes (no tenant context)
	adminApiRoute := r.Group("/admin")
	adminApiRoute.Use(middleware.AuthMiddleware())
	{
		adminApiRoute.POST("/tenants/onboard", tenantHandler.OnboardTenant)
		adminApiRoute.GET("/tenants/:id", tenantHandler.GetTenant)
		adminApiRoute.GET("/tenants/domain/:domain", tenantHandler.GetTenantByDomain)
	}
}
