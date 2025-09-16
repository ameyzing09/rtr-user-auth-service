package main

import (
	"rtr-user-auth-service/handlers"
	"rtr-user-auth-service/internal/db"
	"rtr-user-auth-service/repositories"
	"rtr-user-auth-service/routes"
	"rtr-user-auth-service/services"

	"github.com/gin-gonic/gin"
)

func main() {
	dbInstance := db.InitDB()

	userRepo := repositories.NewGormUserRepo(dbInstance)
	tenantRepo := repositories.NewGormTenantRepo(dbInstance)
	authService := services.NewAuthService(dbInstance, userRepo, tenantRepo)
	userHandler := handlers.NewUserHandler(authService)

	router := gin.Default()
	routes.RegisterRoutes(router, userHandler)

	router.Run(":8082")
}
