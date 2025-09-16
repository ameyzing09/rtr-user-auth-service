package routes

import (
	"rtr-user-auth-service/handlers"
	"rtr-user-auth-service/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, userHandler *handlers.UserHandler) {
	r.POST("/login", userHandler.Login)
	auth := r.Group("/auth")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/register", userHandler.Register)
		auth.GET("/me", userHandler.GetMe)
		auth.GET("/users", userHandler.ListUsers)
	}
}
