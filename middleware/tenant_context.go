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
	"rtr-user-auth-service/utils"

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
		if env == "" {
			env = "local"
		}
		tenantIDHeader := strings.TrimSpace(c.GetHeader("X-Tenant-Id"))
		domainHeader := strings.TrimSpace(c.GetHeader("X-Tenant-Domain"))
		tsHeader := strings.TrimSpace(c.GetHeader("X-Tenant-Ts"))
		sigHeader := strings.TrimSpace(c.GetHeader("X-Tenant-Sig"))

		utils.Debug("[TenantContext] Processing request: env=%s, tenantID=%s, domain=%s, hasSig=%t",
			env, tenantIDHeader, domainHeader, sigHeader != "")

		var tenant *models.Tenant
		var err error
		var resolvedTenantID string

		if tenantIDHeader != "" || tsHeader != "" || sigHeader != "" {
			utils.Debug("[TenantContext] Using signed tenant context")
			tenant, err = handleSignedTenantContext(c, repo, tenantIDHeader, domainHeader, tsHeader, sigHeader)
			if err != nil {
				utils.Debug("[TenantContext] Signed context failed: %v", err)
				return
			}
			resolvedTenantID = tenant.ID
		} else {
			utils.Debug("[TenantContext] Using unsigned tenant context")
			tenant, err = handleUnsignedTenantContext(c, repo, env, domainHeader)
			if err != nil {
				utils.Debug("[TenantContext] Unsigned context failed: %v", err)
				return
			}
			resolvedTenantID = tenant.ID
		}

		utils.Debug("[TenantContext] Successfully resolved tenant: ID=%s, Name=%s, Domain=%s",
			resolvedTenantID, tenant.Name, tenantDomainValue(tenant))

		c.Set(CtxTenantIDKey, resolvedTenantID)
		c.Set(CtxTenantKey, tenant)
		c.Next()
	}
}

func handleSignedTenantContext(c *gin.Context, repo repositories.TenantRepository, tenantID, domain, ts, sig string) (*models.Tenant, error) {
	utils.Debug("[SignedContext] Validating signed tenant context: tenantID=%s, domain=%s", tenantID, domain)

	if tenantID == "" || ts == "" || sig == "" {
		utils.Debug("[SignedContext] Missing required headers: tenantID=%t, ts=%t, sig=%t",
			tenantID != "", ts != "", sig != "")
		abortWithError(c, http.StatusUnauthorized, "missing tenant signature headers")
		return nil, errAborted
	}

	tsValue, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		utils.Debug("[SignedContext] Invalid timestamp format: %s, error: %v", ts, err)
		abortWithError(c, http.StatusUnauthorized, "invalid tenant timestamp")
		return nil, errAborted
	}

	nowMinutes := time.Now().UTC().Unix() / 60
	if diff := minutesDiff(tsValue, nowMinutes); diff > 2 {
		utils.Debug("[SignedContext] Timestamp expired: diff=%d minutes", diff)
		abortWithError(c, http.StatusUnauthorized, "tenant context expired")
		return nil, errAborted
	}

	if !verifyTenantSignature(tenantID, domain, ts, sig) {
		utils.Debug("[SignedContext] Signature verification failed")
		abortWithError(c, http.StatusUnauthorized, "invalid tenant signature")
		return nil, errAborted
	}

	utils.Debug("[SignedContext] Signature verified, looking up tenant by ID: %s", tenantID)
	tenant, err := findTenantByID(c, repo, tenantID)
	if err != nil {
		return nil, errAborted
	}

	if domain != "" {
		tenantDomain := tenantDomainValue(tenant)
		utils.Debug("[SignedContext] Verifying domain match: expected=%s, actual=%s", domain, tenantDomain)
		if tenantDomain == "" || !strings.EqualFold(domain, tenantDomain) {
			utils.Debug("[SignedContext] Domain mismatch: expected=%s, actual=%s", domain, tenantDomain)
			abortWithError(c, http.StatusForbidden, "tenant domain mismatch")
			return nil, errAborted
		}
	}

	utils.Debug("[SignedContext] Successfully validated signed context for tenant: %s", tenantID)
	return tenant, nil
}

func handleUnsignedTenantContext(c *gin.Context, repo repositories.TenantRepository, env, headerDomain string) (*models.Tenant, error) {
	utils.Debug("[UnsignedContext] Resolving tenant from domain: env=%s, headerDomain=%s", env, headerDomain)

	domain := ""
	if headerDomain != "" && env == "local" {
		utils.Debug("[UnsignedContext] Using header domain in local env: %s", headerDomain)
		domain = headerDomain
	}
	// print domain
	utils.Debug("[UnsignedContext] Resolved domain: %s", domain)
	if domain == "" {
		host := c.Request.Host
		if idx := strings.Index(host, ":"); idx > -1 {
			host = host[:idx]
		}
		domain = host
		utils.Debug("[UnsignedContext] Using request host as domain: %s", domain)
	}

	domain = strings.TrimSpace(domain)
	if domain == "" {
		utils.Debug("[UnsignedContext] No domain resolved, aborting")
		abortWithError(c, http.StatusBadRequest, "missing tenant context")
		return nil, errAborted
	}

	utils.Debug("[UnsignedContext] Looking up tenant by domain: %s", domain)
	tenant, err := findTenantByDomain(c, repo, domain)
	if err != nil {
		return nil, errAborted
	}

	utils.Debug("[UnsignedContext] Successfully resolved tenant from domain: %s -> %s", domain, tenant.ID)
	return tenant, nil
}

func findTenantByID(c *gin.Context, repo repositories.TenantRepository, tenantID string) (*models.Tenant, error) {
	if tenant := cacheGetByID(tenantID); tenant != nil {
		utils.Debug("[Cache] Cache HIT for tenant ID: %s", tenantID)
		return tenant, nil
	}

	utils.Debug("[Cache] Cache MISS for tenant ID: %s, querying database", tenantID)
	tenant, err := repo.FindByID(c.Request.Context(), tenantID)
	if err != nil {
		utils.Debug("[DB] Tenant lookup failed for ID %s: %v", tenantID, err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			abortWithError(c, http.StatusNotFound, "tenant not found")
		} else {
			abortWithError(c, http.StatusInternalServerError, "tenant lookup failed")
		}
		return nil, errAborted
	}
	if tenant == nil {
		utils.Debug("[DB] Tenant not found for ID: %s", tenantID)
		abortWithError(c, http.StatusNotFound, "tenant not found")
		return nil, errAborted
	}

	utils.Debug("[DB] Found tenant by ID: %s -> %s", tenantID, tenant.Name)
	cacheStore(tenant)
	return tenant, nil
}

func findTenantByDomain(c *gin.Context, repo repositories.TenantRepository, domain string) (*models.Tenant, error) {
	if tenant := cacheGetByDomain(domain); tenant != nil {
		utils.Debug("[Cache] Cache HIT for domain: %s", domain)
		return tenant, nil
	}

	utils.Debug("[Cache] Cache MISS for domain: %s, querying database", domain)
	tenant, err := repo.FindByDomain(c.Request.Context(), domain)
	if err != nil {
		utils.Debug("[DB] Tenant lookup failed for domain %s: %v", domain, err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			abortWithError(c, http.StatusNotFound, "tenant not found")
		} else {
			abortWithError(c, http.StatusInternalServerError, "tenant lookup failed")
		}
		return nil, errAborted
	}
	if tenant == nil {
		utils.Debug("[DB] Tenant not found for domain: %s", domain)
		abortWithError(c, http.StatusNotFound, "tenant not found")
		return nil, errAborted
	}

	utils.Debug("[DB] Found tenant by domain: %s -> %s (%s)", domain, tenant.ID, tenant.Name)
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
	if d := tenantDomainValue(tenant); d != "" {
		tcCache.byDomain[strings.ToLower(d)] = entry
		utils.Debug("[Cache] Stored tenant in cache: ID=%s, Domain=%s, TTL=%v", tenant.ID, d, cacheTTL)
	} else {
		utils.Debug("[Cache] Stored tenant in cache: ID=%s, Domain=<none>, TTL=%v", tenant.ID, cacheTTL)
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
	// Do not log secrets to avoid leaking sensitive information
	// fmt.Printf("Verifying signature with secrets: %v\n", secrets)
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

func tenantDomainValue(t *models.Tenant) string {
	if t == nil || t.Domain == nil {
		return ""
	}
	return *t.Domain
}
