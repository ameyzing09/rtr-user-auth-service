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

	// Protected tenant routes (with auth first, then authenticated tenant context, and CSRF protection)
	// AuthenticatedTenantContext derives tenant from the authenticated user's JWT token
	// This eliminates the need for client-side tenant secrets
	protectedRoute := r.Group("/")
	protectedRoute.Use(middleware.AuthMiddleware(), middleware.AuthenticatedTenantContext(tenantRepo), middleware.CSRFProtection())
	{
		// Profile routes - all authenticated users can access their own profile
		protectedRoute.GET("/me", userHandler.GetMe)
		protectedRoute.POST("/me/change-password", userHandler.ChangePassword)
		protectedRoute.POST("/logout", userHandler.Logout)

		// User management - requires member permissions
		protectedRoute.GET("/users", middleware.RequirePermission(string(models.PermMemberAll)), userHandler.ListUsers)
		protectedRoute.POST("/users", middleware.RequirePermission(string(models.PermMemberAll)), userHandler.CreateUser)

		// Tenant settings - requires settings permissions
		tenantScope := protectedRoute.Group("/tenant")
		tenantScope.PUT("/settings", middleware.RequirePermission(string(models.PermSettingsAll)), tenantSettingHandler.Put)
	}

	// Superadmin control-plane routes (no tenant context, with CSRF protection)
	admin := r.Group("/")
	admin.Use(middleware.ControlPlaneScope(), middleware.AuthMiddleware(), middleware.CSRFProtection())
	{
		admin.POST("/admin/logout", userHandler.Logout)

		// Tenant CRUD operations
		admin.GET("/admin/tenants", middleware.RequirePermission(string(models.PermTenantList)), tenantAdminHandler.List)
		admin.POST("/admin/tenant/create", middleware.RequirePermission(string(models.PermTenantCreate)), tenantAdminHandler.Create)
		admin.GET("/admin/tenant/:id", middleware.RequirePermission(string(models.PermTenantRead)), tenantAdminHandler.Get)
		admin.PUT("/admin/tenant/:id", middleware.RequirePermission(string(models.PermTenantUpdate)), tenantAdminHandler.Update)
		admin.DELETE("/admin/tenant/:id", middleware.RequirePermission(string(models.PermTenantUpdate)), tenantAdminHandler.Delete)
		admin.GET("/tenant/:id/status", middleware.RequirePermission(string(models.PermTenantStatus)), tenantAdminHandler.Status)
		admin.POST("/tenant/:id/retry", middleware.RequirePermission(string(models.PermTenantStatus)), tenantAdminHandler.Retry)

		// Tenant archive operations
		admin.GET("/admin/tenants/archived", middleware.RequirePermission(string(models.PermTenantList)), tenantAdminHandler.ListArchived)
		admin.GET("/admin/tenant/:id/archived", middleware.RequirePermission(string(models.PermTenantRead)), tenantAdminHandler.GetArchived)

		// Subscription management - requires tenant:update permission
		admin.GET("/admin/tenant/:id/subscription", middleware.RequirePermission(string(models.PermTenantRead)), subscriptionAdminHandler.Get)
		admin.POST("/admin/tenant/:id/subscription/activate", middleware.RequirePermission(string(models.PermTenantUpdate)), subscriptionAdminHandler.Activate)
		admin.POST("/admin/tenant/:id/subscription/suspend", middleware.RequirePermission(string(models.PermTenantUpdate)), subscriptionAdminHandler.Suspend)
		admin.POST("/admin/tenant/:id/subscription/resume", middleware.RequirePermission(string(models.PermTenantUpdate)), subscriptionAdminHandler.Resume)
		admin.POST("/admin/tenant/:id/subscription/cancel", middleware.RequirePermission(string(models.PermTenantUpdate)), subscriptionAdminHandler.Cancel)
		admin.POST("/admin/change-password", middleware.RequirePermission(string(models.PermTenantUpdate)), userHandler.SuperadminChangePassword)
	}
}
 