package main

import (
	"log"
	"os"
	"rtr-user-auth-service/handlers"
	"rtr-user-auth-service/internal/db"
	"rtr-user-auth-service/middleware"
	"rtr-user-auth-service/repositories"
	"rtr-user-auth-service/routes"
	"rtr-user-auth-service/services"

	"github.com/gin-gonic/gin"
)

func main() {
	if gm := os.Getenv("GIN_MODE"); gm != "" {
		gin.SetMode(gm)
	}

	dbInstance := db.InitDB()

	userRepo := repositories.NewGormUserRepo(dbInstance)
	tenantRepo := repositories.NewGormTenantRepo(dbInstance)
	tenantSettingRepo := repositories.NewGormTenantSettingRepo(dbInstance)

	authService := services.NewAuthService(dbInstance, userRepo, tenantRepo)
	tenantSettingService := services.NewTenantSettingService(tenantSettingRepo)
	tenantService := services.NewTenantService(dbInstance, tenantRepo, userRepo, tenantSettingRepo)

	userHandler := handlers.NewUserHandler(authService)
	tenantSettingHandler := handlers.NewTenantSettingHandler(tenantSettingService)
	tenantHandler := handlers.NewTenantHandler(tenantService)

	router := gin.New()
	router.Use(gin.Recovery(), middleware.CORS())

	routes.RegisterRoutes(router, userHandler, tenantSettingHandler, tenantHandler, tenantRepo)

	if err := router.Run(":8082"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
