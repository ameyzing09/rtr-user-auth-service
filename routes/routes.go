package routes

import (
	"net/http"
	"rtr-user-auth-service/handlers"
	"rtr-user-auth-service/middleware"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/repositories"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, userHandler *handlers.UserHandler, tenantSettingHandler *handlers.TenantSettingHandler, tenantAdminHandler *handlers.TenantCreateHandler, subscriptionAdminHandler *handlers.SubscriptionAdminHandler, tenantRepo repositories.TenantRepository) {
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "rtr-user-auth-service: ok")
	})

	r.POST("/admin/login", userHandler.AdminLogin)

	publicRoute := r.Group("/")
	publicRoute.Use(middleware.TenantContext(tenantRepo))
	{
		publicRoute.POST("/login", userHandler.Login)

		tenantScope := publicRoute.Group("/tenant")
		tenantScope.GET("/settings", tenantSettingHandler.Get)
	}

	// Protected tenant routes (with tenant context and auth)
	protectedRoute := r.Group("/")
	protectedRoute.Use(middleware.TenantContext(tenantRepo), middleware.AuthMiddleware())
	{
		// Profile routes - all authenticated users can access their own profile
		protectedRoute.GET("/me", userHandler.GetMe)
		protectedRoute.POST("/me/change-password", userHandler.ChangePassword)
		protectedRoute.POST("/logout", userHandler.Logout)

		// User management - ADMIN and HR can manage users
		protectedRoute.GET("/users", middleware.RequireAny(models.RoleAdmin, models.RoleHR), userHandler.ListUsers)
		protectedRoute.POST("/users", middleware.RequireAny(models.RoleAdmin, models.RoleHR), userHandler.CreateUser)

		// Tenant settings - only ADMIN can update tenant settings
		tenantScope := protectedRoute.Group("/tenant")
		tenantScope.PUT("/settings", middleware.RequireRole(models.RoleAdmin), tenantSettingHandler.Put)
	}

	// Superadmin control-plane routes (no tenant context)
	admin := r.Group("/")
	admin.Use(middleware.ControlPlaneScope(), middleware.AuthMiddleware())
	{
		admin.POST("/admin/logout", userHandler.Logout)

		// Tenant CRUD operations
		admin.GET("/admin/tenants", middleware.RequireRole(models.RoleSuperAdmin), tenantAdminHandler.List)
		admin.POST("/admin/tenant/create", middleware.RequireRole(models.RoleSuperAdmin), tenantAdminHandler.Create)
		admin.GET("/admin/tenant/:id", middleware.RequireRole(models.RoleSuperAdmin), tenantAdminHandler.Get)
		admin.PUT("/admin/tenant/:id", middleware.RequireRole(models.RoleSuperAdmin), tenantAdminHandler.Update)
		admin.DELETE("/admin/tenant/:id", middleware.RequireRole(models.RoleSuperAdmin), tenantAdminHandler.Delete)
		admin.GET("/tenant/:id/status", middleware.RequireRole(models.RoleSuperAdmin), tenantAdminHandler.Status)
		admin.POST("/tenant/:id/retry", middleware.RequireRole(models.RoleSuperAdmin), tenantAdminHandler.Retry)

		// Subscription management
		admin.GET("/admin/tenant/:id/subscription", middleware.RequireRole(models.RoleSuperAdmin), subscriptionAdminHandler.Get)
		admin.POST("/admin/tenant/:id/subscription/activate", middleware.RequireRole(models.RoleSuperAdmin), subscriptionAdminHandler.Activate)
		admin.POST("/admin/tenant/:id/subscription/suspend", middleware.RequireRole(models.RoleSuperAdmin), subscriptionAdminHandler.Suspend)
		admin.POST("/admin/tenant/:id/subscription/resume", middleware.RequireRole(models.RoleSuperAdmin), subscriptionAdminHandler.Resume)
		admin.POST("/admin/tenant/:id/subscription/cancel", middleware.RequireRole(models.RoleSuperAdmin), subscriptionAdminHandler.Cancel)
		admin.POST("/admin/change-password", middleware.RequireRole(models.RoleSuperAdmin), userHandler.SuperadminChangePassword)
	}
}
