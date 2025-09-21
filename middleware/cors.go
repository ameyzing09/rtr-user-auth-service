package middleware

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	allowHeaders := []string{"Authorization", "Content-Type"}
	if env == "local" {
		allowHeaders = append(allowHeaders, "X-Tenant-Domain")
	}

	cfg := cors.Config{
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     allowHeaders,
		AllowOriginFunc: func(origin string) bool {
			return strings.HasSuffix(origin, ".recrutr.in") || strings.HasSuffix(origin, "http://localhost:") || strings.HasSuffix(origin, "http://127.0.0.1:")
		},
	}
	return cors.New(cfg)
}
