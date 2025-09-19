package middleware

import (
	"net/http"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/repositories"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	CtxTenantIDKey = "tenantID"
	CtxTenantKey   = "tenant"
)

type cachedTenant struct {
	t   *models.Tenant
	exp time.Time
}

var (
	trCache   = map[string]cachedTenant{}
	trCacheMu sync.RWMutex
	cacheTTL  = 30 * time.Second
)

func TenantResolver(tenantRepo repositories.TenantRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.GetHeader("x-tenant-domain")
		if host == "" {
			host = c.Request.Host
		}

		if i := strings.Index(host, ":"); i > -1 {
			host = host[:i]
		}
		if host == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing host"})
		}

		trCacheMu.RLock()
		if ct, ok := trCache[host]; ok && time.Now().Before(ct.exp) {
			trCacheMu.RUnlock()
			c.Set(CtxTenantIDKey, ct.t.ID)
			c.Set(CtxTenantKey, ct.t)
			c.Next()
			return
		}
		trCacheMu.RUnlock()

		t, err := tenantRepo.FindByDomain(c, host)
		if err != nil || t == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "unknown tenant domain"})
			return
		}

		trCacheMu.Lock()
		trCache[host] = cachedTenant{t: t, exp: time.Now().Add(cacheTTL)}
		trCacheMu.Unlock()

		c.Set(CtxTenantIDKey, t.ID)
		c.Set(CtxTenantKey, t)
		c.Next()
	}
}
