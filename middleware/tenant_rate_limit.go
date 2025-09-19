package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type tenantLimiter struct {
	mu    sync.Mutex
	pool  map[string]*rate.Limiter
	rps   rate.Limit
	burst int
}

func TenantRateLimiter(rps float64, burst int) gin.HandlerFunc {
	tl := &tenantLimiter{pool: map[string]*rate.Limiter{}, rps: rate.Limit(rps), burst: burst}
	return func(c *gin.Context) {
		tid := c.GetString(CtxTenantIDKey)
		if tid == "" {
			c.Next()
			return
		}
		tl.mu.Lock()
		limiter, ok := tl.pool[tid]
		if !ok {
			limiter = rate.NewLimiter(tl.rps, tl.burst)
			tl.pool[tid] = limiter
		}
		tl.mu.Unlock()
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}
		c.Next()
	}
}
