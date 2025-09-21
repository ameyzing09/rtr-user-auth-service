package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"rtr-user-auth-service/models"
	"rtr-user-auth-service/repositories"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type cachedTenant struct {
	tenant *models.Tenant
	exp    time.Time
}

type tenantCache struct {
	mu       sync.RWMutex
	byID     map[string]cachedTenant
	byDomain map[string]cachedTenant
}

var (
	tcCache = tenantCache{
		byID:     map[string]cachedTenant{},
		byDomain: map[string]cachedTenant{},
	}
	cacheTTL   = 30 * time.Second
	errAborted = errors.New("aborted")
)

func TenantContext(repo repositories.TenantRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))

		tenantIDHeader := strings.TrimSpace(c.GetHeader("X-Tenant-Id"))
		domainHeader := strings.TrimSpace(c.GetHeader("X-Tenant-Domain"))
		tsHeader := strings.TrimSpace(c.GetHeader("X-Tenant-Ts"))
		sigHeader := strings.TrimSpace(c.GetHeader("X-Tenant-Sig"))

		var tenant *models.Tenant
		var err error
		var resolvedTenantID string

		if tenantIDHeader != "" || tsHeader != "" || sigHeader != "" {
			tenant, err = handleSignedTenantContext(c, repo, tenantIDHeader, domainHeader, tsHeader, sigHeader)
			if err != nil {
				return
			}
			resolvedTenantID = tenant.ID
		} else {
			tenant, err = handleUnsignedTenantContext(c, repo, env, domainHeader)
			if err != nil {
				return
			}
			resolvedTenantID = tenant.ID
		}

		c.Set(CtxTenantIDKey, resolvedTenantID)
		c.Set(CtxTenantKey, tenant)
		c.Next()
	}
}

func handleSignedTenantContext(c *gin.Context, repo repositories.TenantRepository, tenantID, domain, ts, sig string) (*models.Tenant, error) {
	if tenantID == "" || ts == "" || sig == "" {
		abortWithError(c, http.StatusUnauthorized, "missing tenant signature headers")
		return nil, errAborted
	}

	tsValue, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		abortWithError(c, http.StatusUnauthorized, "invalid tenant timestamp")
		return nil, errAborted
	}

	nowMinutes := time.Now().UTC().Unix() / 60
	if diff := minutesDiff(tsValue, nowMinutes); diff > 2 {
		abortWithError(c, http.StatusUnauthorized, "tenant context expired")
		return nil, errAborted
	}

	if !verifyTenantSignature(tenantID, domain, ts, sig) {
		abortWithError(c, http.StatusUnauthorized, "invalid tenant signature")
		return nil, errAborted
	}

	tenant, err := findTenantByID(c, repo, tenantID)
	if err != nil {
		return nil, errAborted
	}

	if domain != "" && !strings.EqualFold(domain, tenant.Domain) {
		abortWithError(c, http.StatusForbidden, "tenant domain mismatch")
		return nil, errAborted
	}

	return tenant, nil
}

func handleUnsignedTenantContext(c *gin.Context, repo repositories.TenantRepository, env, headerDomain string) (*models.Tenant, error) {
	domain := ""
	if headerDomain != "" && env == "local" {
		domain = headerDomain
	}

	if domain == "" {
		host := c.Request.Host
		if idx := strings.Index(host, ":"); idx > -1 {
			host = host[:idx]
		}
		domain = host
	}

	domain = strings.TrimSpace(domain)
	if domain == "" {
		abortWithError(c, http.StatusBadRequest, "missing tenant context")
		return nil, errAborted
	}

	tenant, err := findTenantByDomain(c, repo, domain)
	if err != nil {
		return nil, errAborted
	}

	return tenant, nil
}

func findTenantByID(c *gin.Context, repo repositories.TenantRepository, tenantID string) (*models.Tenant, error) {
	if tenant := cacheGetByID(tenantID); tenant != nil {
		return tenant, nil
	}

	tenant, err := repo.FindByID(c.Request.Context(), tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			abortWithError(c, http.StatusNotFound, "tenant not found")
		} else {
			abortWithError(c, http.StatusInternalServerError, "tenant lookup failed")
		}
		return nil, errAborted
	}
	if tenant == nil {
		abortWithError(c, http.StatusNotFound, "tenant not found")
		return nil, errAborted
	}

	cacheStore(tenant)
	return tenant, nil
}

func findTenantByDomain(c *gin.Context, repo repositories.TenantRepository, domain string) (*models.Tenant, error) {
	if tenant := cacheGetByDomain(domain); tenant != nil {
		return tenant, nil
	}

	tenant, err := repo.FindByDomain(c.Request.Context(), domain)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			abortWithError(c, http.StatusNotFound, "tenant not found")
		} else {
			abortWithError(c, http.StatusInternalServerError, "tenant lookup failed")
		}
		return nil, errAborted
	}
	if tenant == nil {
		abortWithError(c, http.StatusNotFound, "tenant not found")
		return nil, errAborted
	}

	cacheStore(tenant)
	return tenant, nil
}

func cacheGetByID(id string) *models.Tenant {
	key := strings.ToLower(id)
	tcCache.mu.RLock()
	entry, ok := tcCache.byID[key]
	tcCache.mu.RUnlock()
	if !ok || time.Now().After(entry.exp) {
		if ok {
			tcCache.mu.Lock()
			delete(tcCache.byID, key)
			tcCache.mu.Unlock()
		}
		return nil
	}
	return entry.tenant
}

func cacheGetByDomain(domain string) *models.Tenant {
	key := strings.ToLower(domain)
	tcCache.mu.RLock()
	entry, ok := tcCache.byDomain[key]
	tcCache.mu.RUnlock()
	if !ok || time.Now().After(entry.exp) {
		if ok {
			tcCache.mu.Lock()
			delete(tcCache.byDomain, key)
			tcCache.mu.Unlock()
		}
		return nil
	}
	return entry.tenant
}

func cacheStore(tenant *models.Tenant) {
	if tenant == nil {
		return
	}
	entry := cachedTenant{tenant: tenant, exp: time.Now().Add(cacheTTL)}
	tcCache.mu.Lock()
	tcCache.byID[strings.ToLower(tenant.ID)] = entry
	if tenant.Domain != "" {
		tcCache.byDomain[strings.ToLower(tenant.Domain)] = entry
	}
	tcCache.mu.Unlock()
}

func minutesDiff(a, b int64) int64 {
	if a > b {
		return a - b
	}
	return b - a
}

func verifyTenantSignature(tenantID, domain, ts, sig string) bool {
	secrets := []string{strings.TrimSpace(os.Getenv("TENANT_CTX_SECRET")), strings.TrimSpace(os.Getenv("TENANT_CTX_SECRET_PREV"))}
	payload := tenantID + "." + domain + "." + ts
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(payload))
		expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		if hmac.Equal([]byte(expected), []byte(sig)) {
			return true
		}
	}
	return false
}

func abortWithError(c *gin.Context, status int, msg string) {
	if msg == "" {
		msg = http.StatusText(status)
	}
	c.AbortWithStatusJSON(status, gin.H{"error": msg})
}
