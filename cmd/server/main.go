package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ameyzing09/rtr-user-auth-service/internal/config"
	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/repositories"
	"github.com/ameyzing09/rtr-user-auth-service/internal/handlers"
	"github.com/ameyzing09/rtr-user-auth-service/internal/middleware"
	"github.com/ameyzing09/rtr-user-auth-service/internal/services"
	"github.com/ameyzing09/rtr-user-auth-service/internal/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// @title           Recrutr Auth Service API
// @version         1.0
// @description     Multi-tenant authentication service for Recrutr platform
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize database
	db, err := config.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Check database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize repositories
	tenantRepo := repositories.NewTenantRepository(db.DB)
	userRepo := repositories.NewUserRepository(db.DB)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db.DB)

	// Initialize JWT service
	jwtService := utils.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)

	// Initialize services
	authService := services.NewAuthService(userRepo, refreshTokenRepo, tenantRepo, jwtService)
	userService := services.NewUserService(userRepo, tenantRepo)
	tenantService := services.NewTenantService(tenantRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	tenantHandler := handlers.NewTenantHandler(tenantService)

	// Initialize Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(middleware.ErrorHandler())

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.CORS.AllowedOrigins
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(corsConfig))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"service":   "recrutr-auth-service",
		})
	})

	// API routes
	api := router.Group("/api/v1")

	// Public routes
	public := api.Group("")
	{
		// Auth routes
		auth := public.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Public tenant routes
		public.GET("/tenants/by-domain", tenantHandler.GetTenantByDomain)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		// Auth routes (protected)
		auth := protected.Group("/auth")
		{
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/profile", userHandler.GetProfile)
		}

		// Tenant routes (admin only)
		tenants := protected.Group("/tenants")
		tenants.Use(middleware.RequireAdmin())
		{
			tenants.POST("", tenantHandler.CreateTenant)
			tenants.GET("", tenantHandler.ListTenants)
			tenants.GET("/:tenantId", tenantHandler.GetTenant)
			tenants.PUT("/:tenantId", tenantHandler.UpdateTenant)
			tenants.DELETE("/:tenantId", tenantHandler.DeleteTenant)
		}

		// Tenant-scoped user routes
		tenantUsers := protected.Group("/tenants/:tenantId/users")
		tenantUsers.Use(middleware.RequireSameTenantOrAdmin())
		{
			// Create user (Admin or HR)
			tenantUsers.POST("", middleware.RequireAdminOrHR(), userHandler.CreateUser)
			
			// List users (Admin or HR)
			tenantUsers.GET("", middleware.RequireAdminOrHR(), userHandler.ListUsers)
			
			// Individual user operations
			userRoutes := tenantUsers.Group("/:userId")
			userRoutes.Use(middleware.RequireSameUserOrAdminOrHR())
			{
				userRoutes.GET("", userHandler.GetUser)
				userRoutes.PUT("", userHandler.UpdateUser)
				userRoutes.DELETE("", middleware.RequireAdminOrHR(), userHandler.DeleteUser)
			}
		}
	}

	// 404 handler
	router.NoRoute(middleware.NotFoundHandler())

	// 405 handler
	router.NoMethod(middleware.MethodNotAllowedHandler())

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}