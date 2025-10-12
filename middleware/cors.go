package middleware

import (
	"net/url"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	if env == "" {
		env = "local"
	}

	// X-CSRF-Token is always allowed for CSRF protection
	allowHeaders := []string{"Authorization", "Content-Type", "Idempotency-Key", "X-CSRF-Token"}
	if env == "local" {
		allowHeaders = append(allowHeaders, "X-Tenant-Domain", "X-Tenant-ID", "X-Tenant-Ts", "X-Tenant-Sig")
	}

	cfg := cors.Config{
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     allowHeaders,
		AllowOriginFunc: func(origin string) bool {
			if origin == "" {
				return false
			}

			parsed, err := url.Parse(origin)
			if err != nil {
				return false
			}

			host := parsed.Hostname()
			if host == "" {
				return false
			}

			switch env {
			case "local", "dev":
				if strings.EqualFold(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
					return true
				}
			}

			if strings.EqualFold(host, "tenants.recrutr.in") {
				return true
			}
			if strings.HasSuffix(host, ".tenants.recrutr.in") {
				return true
			}

			// Allow shared admin domains
			if strings.EqualFold(host, "admin.recrutr.in") || strings.HasSuffix(host, ".admin.recrutr.in") {
				return true
			}

			return false
		},
	}

	return cors.New(cfg)
}
