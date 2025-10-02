package main

import (
	"fmt"
	"log"

	"rtr-user-auth-service/config"
	"rtr-user-auth-service/handlers"
	"rtr-user-auth-service/internal/db"
	"rtr-user-auth-service/middleware"
	"rtr-user-auth-service/repositories"
	"rtr-user-auth-service/routes"
	"rtr-user-auth-service/services"
	"rtr-user-auth-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Set Gin mode
	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	// Initialize logger (logger is automatically initialized via init())
	utils.Info("Starting application with environment: %s", cfg.Server.Env)

	// Initialize database
	dbInstance, err := db.InitDB(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	utils.Info("Database connection established")

	// Initialize repositories
	repos := initRepositories(dbInstance)

	// Initialize services
	svcs := initServices(dbInstance, repos, cfg)

	// Initialize handlers
	hndlrs := initHandlers(svcs)

	// Setup router
	router := setupRouter(hndlrs, repos.TenantRepo, cfg)

	// Start server
	addr := ":" + cfg.Server.Port
	utils.Info("Server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

type repositoriesContainer struct {
	UserRepo          services.UserRepository
	TenantRepo        services.TenantRepository
	TenantSettingRepo services.TenantSettingRepository
	IdempotencyRepo   services.IdempotencyRepository
}

func initRepositories(db *gorm.DB) *repositoriesContainer {
	return &repositoriesContainer{
		UserRepo:          repositories.NewGormUserRepo(db),
		TenantRepo:        repositories.NewGormTenantRepo(db),
		TenantSettingRepo: repositories.NewGormTenantSettingRepo(db),
		IdempotencyRepo:   repositories.NewGormIdempotencyRepo(db),
	}
}

type servicesContainer struct {
	AuthService          services.AuthService
	TenantSettingService services.TenantSettingService
	TenantService        services.TenantService
}

func initServices(db *gorm.DB, repos *repositoriesContainer, cfg *config.Config) *servicesContainer {
	return &servicesContainer{
		AuthService:          services.NewAuthService(db, repos.UserRepo, repos.TenantRepo),
		TenantSettingService: services.NewTenantSettingService(repos.TenantSettingRepo),
		TenantService:        services.NewTenantService(db, repos.TenantRepo, repos.IdempotencyRepo),
	}
}

type handlersContainer struct {
	UserHandler          *handlers.UserHandler
	TenantSettingHandler *handlers.TenantSettingHandler
	TenantAdminHandler   *handlers.TenantCreateHandler
}

func initHandlers(svcs *servicesContainer) *handlersContainer {
	return &handlersContainer{
		UserHandler:          handlers.NewUserHandler(svcs.AuthService),
		TenantSettingHandler: handlers.NewTenantSettingHandler(svcs.TenantSettingService),
		TenantAdminHandler:   handlers.NewTenantCreateHandler(svcs.TenantService),
	}
}

func setupRouter(hndlrs *handlersContainer, tenantRepo services.TenantRepository, cfg *config.Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	routes.RegisterRoutes(
		router,
		hndlrs.UserHandler,
		hndlrs.TenantSettingHandler,
		hndlrs.TenantAdminHandler,
		tenantRepo,
	)

	return router
}
