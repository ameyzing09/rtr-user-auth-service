package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/semaphore"
)

func TenantConcurrency(maxInFlight int64) gin.HandlerFunc {
	var mu sync.Mutex
	gates := map[string]*semaphore.Weighted{}
	return func(c *gin.Context) {
		tid := c.GetString(CtxTenantIDKey)
		if tid == "" {
			c.Next()
			return
		}

		mu.Lock()
		sem, ok := gates[tid]
		if !ok {
			sem = semaphore.NewWeighted(maxInFlight)
			gates[tid] = sem
		}
		mu.Unlock()

		if !sem.TryAcquire(1) {
			c.AbortWithStatusJSON(429, gin.H{"error": "Too many concurrent requests for this tenant"})
			return
		}
		defer sem.Release(1)
		c.Next()
	}

}
