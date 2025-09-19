package middleware

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	cfg := cors.Config{
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Tenant-Domain"},
		AllowOriginFunc: func(origin string) bool {
			return strings.HasSuffix(origin, ".recrutr.in") || strings.HasSuffix(origin, "http://localhost:") || strings.HasSuffix(origin, "http://127.0.0.1:")
		},
	}
	return cors.New(cfg)
}
