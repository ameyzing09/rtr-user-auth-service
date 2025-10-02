package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"rtr-user-auth-service/models"
	"rtr-user-auth-service/repositories"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type stubTenantRepo struct {
	tenants map[string]*models.Tenant
}

var _ repositories.TenantRepository = (*stubTenantRepo)(nil)

func strPtr(value string) *string {
	return &value
}

func (s *stubTenantRepo) Create(ctx context.Context, tenant *models.Tenant) error {
	if s.tenants == nil {
		s.tenants = make(map[string]*models.Tenant)
	}
	copy := *tenant
	s.tenants[tenant.ID] = &copy
	return nil
}

func (s *stubTenantRepo) Update(ctx context.Context, tenant *models.Tenant) error {
	if s.tenants == nil {
		s.tenants = make(map[string]*models.Tenant)
	}
	copy := *tenant
	s.tenants[tenant.ID] = &copy
	return nil
}

func (s *stubTenantRepo) UpdateStatus(ctx context.Context, tenantID string, status models.TenantStatus) error {
	if s.tenants == nil {
		return gorm.ErrRecordNotFound
	}
	tenant, exists := s.tenants[tenantID]
	if !exists {
		return gorm.ErrRecordNotFound
	}
	tenant.Status = status
	return nil
}

func (s *stubTenantRepo) UpdateStatusWithReason(ctx context.Context, tenantID string, status models.TenantStatus, reason string) error {
	if s.tenants == nil {
		return gorm.ErrRecordNotFound
	}
	tenant, exists := s.tenants[tenantID]
	if !exists {
		return gorm.ErrRecordNotFound
	}
	tenant.Status = status
	if reason != "" {
		tenant.FailedReason = &reason
	}
	return nil
}

func (s *stubTenantRepo) FindBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	for _, tenant := range s.tenants {
		if tenant.Slug != nil && *tenant.Slug == slug {
			return tenant, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubTenantRepo) FindByDomain(ctx context.Context, domain string) (*models.Tenant, error) {
	for _, tenant := range s.tenants {
		if tenant.Domain != nil && *tenant.Domain == domain {
			return tenant, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubTenantRepo) FindByID(ctx context.Context, tenantID string) (*models.Tenant, error) {
	tenant, ok := s.tenants[tenantID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return tenant, nil
}

func (s *stubTenantRepo) ListAll(ctx context.Context) ([]models.Tenant, error) {
	items := make([]models.Tenant, 0, len(s.tenants))
	for _, tenant := range s.tenants {
		copy := *tenant
		items = append(items, copy)
	}
	return items, nil
}

func init() {
	gin.SetMode(gin.TestMode)
}

func resetTenantCache() {
	tcCache.mu.Lock()
	tcCache.byID = map[string]cachedTenant{}
	tcCache.byDomain = map[string]cachedTenant{}
	tcCache.mu.Unlock()
}

func signedHeaders(tenantID, domain, secret string, ts time.Time) (timestamp string, sig string) {
	minutes := ts.UTC().Unix() / 60
	timestamp = strconv.FormatInt(minutes, 10)
	payload := tenantID + "." + domain + "." + timestamp
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig = base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return
}

func TestTenantContext_SignedHeadersCurrentSecret(t *testing.T) {
	resetTenantCache()

	repo := &stubTenantRepo{tenants: map[string]*models.Tenant{
		"tenant-1": {ID: "tenant-1", Domain: strPtr("acme.test")},
	}}

	t.Setenv("TENANT_CTX_SECRET", "secret-current")
	t.Setenv("TENANT_CTX_SECRET_PREV", "")
	t.Setenv("ENV", "prod")

	ts, sig := signedHeaders("tenant-1", "acme.test", "secret-current", time.Now())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-Id", "tenant-1")
	req.Header.Set("X-Tenant-Domain", "acme.test")
	req.Header.Set("X-Tenant-Ts", ts)
	req.Header.Set("X-Tenant-Sig", sig)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	TenantContext(repo)(c)

	if c.IsAborted() {
		t.Fatalf("expected request to continue, got status %d", w.Code)
	}
	if got := c.GetString(CtxTenantIDKey); got != "tenant-1" {
		t.Fatalf("expected tenant id tenant-1, got %s", got)
	}
	tenantVal, exists := c.Get(CtxTenantKey)
	if !exists {
		t.Fatalf("tenant missing in context")
	}
	tenant, ok := tenantVal.(*models.Tenant)
	if !ok || tenant.Domain == nil || *tenant.Domain != "acme.test" {
		t.Fatalf("unexpected tenant payload: %#v", tenantVal)
	}
}

func TestTenantContext_SignedHeadersPreviousSecret(t *testing.T) {
	resetTenantCache()

	repo := &stubTenantRepo{tenants: map[string]*models.Tenant{
		"tenant-1": {ID: "tenant-1", Domain: strPtr("acme.test")},
	}}

	t.Setenv("TENANT_CTX_SECRET", "secret-current")
	t.Setenv("TENANT_CTX_SECRET_PREV", "secret-prev")
	t.Setenv("ENV", "prod")

	ts, sig := signedHeaders("tenant-1", "acme.test", "secret-prev", time.Now())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-Id", "tenant-1")
	req.Header.Set("X-Tenant-Domain", "acme.test")
	req.Header.Set("X-Tenant-Ts", ts)
	req.Header.Set("X-Tenant-Sig", sig)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	TenantContext(repo)(c)

	if c.IsAborted() {
		t.Fatalf("expected request to continue, got status %d", w.Code)
	}
}

func TestTenantContext_InvalidSignature(t *testing.T) {
	resetTenantCache()

	repo := &stubTenantRepo{tenants: map[string]*models.Tenant{
		"tenant-1": {ID: "tenant-1", Domain: strPtr("acme.test")},
	}}

	t.Setenv("TENANT_CTX_SECRET", "secret-current")
	t.Setenv("TENANT_CTX_SECRET_PREV", "")
	t.Setenv("ENV", "prod")

	ts, sig := signedHeaders("tenant-1", "acme.test", "other-secret", time.Now())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-Id", "tenant-1")
	req.Header.Set("X-Tenant-Domain", "acme.test")
	req.Header.Set("X-Tenant-Ts", ts)
	req.Header.Set("X-Tenant-Sig", sig)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	TenantContext(repo)(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestTenantContext_ExpiredSignature(t *testing.T) {
	resetTenantCache()

	repo := &stubTenantRepo{tenants: map[string]*models.Tenant{
		"tenant-1": {ID: "tenant-1", Domain: strPtr("acme.test")},
	}}

	t.Setenv("TENANT_CTX_SECRET", "secret-current")
	t.Setenv("TENANT_CTX_SECRET_PREV", "")
	t.Setenv("ENV", "prod")

	ts, sig := signedHeaders("tenant-1", "acme.test", "secret-current", time.Now().Add(-10*time.Minute))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-Id", "tenant-1")
	req.Header.Set("X-Tenant-Domain", "acme.test")
	req.Header.Set("X-Tenant-Ts", ts)
	req.Header.Set("X-Tenant-Sig", sig)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	TenantContext(repo)(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestTenantContext_HostFallback(t *testing.T) {
	resetTenantCache()

	repo := &stubTenantRepo{tenants: map[string]*models.Tenant{
		"tenant-1": {ID: "tenant-1", Domain: strPtr("acme.test")},
	}}

	t.Setenv("TENANT_CTX_SECRET", "")
	t.Setenv("TENANT_CTX_SECRET_PREV", "")
	t.Setenv("ENV", "prod")

	req := httptest.NewRequest(http.MethodGet, "http://acme.test/resource", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	TenantContext(repo)(c)

	if c.IsAborted() {
		t.Fatalf("expected host fallback to succeed, got %d", w.Code)
	}
	if got := c.GetString(CtxTenantIDKey); got != "tenant-1" {
		t.Fatalf("expected tenant id tenant-1, got %s", got)
	}
}

func TestTenantContext_LocalUnsignedDomain(t *testing.T) {
	resetTenantCache()

	repo := &stubTenantRepo{tenants: map[string]*models.Tenant{
		"tenant-1": {ID: "tenant-1", Domain: strPtr("acme.test")},
	}}

	t.Setenv("TENANT_CTX_SECRET", "")
	t.Setenv("TENANT_CTX_SECRET_PREV", "")
	t.Setenv("ENV", "local")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-Domain", "acme.test")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	TenantContext(repo)(c)

	if c.IsAborted() {
		t.Fatalf("expected local unsigned domain to be accepted, got %d", w.Code)
	}
}

func TestTenantContext_MismatchedDomain(t *testing.T) {
	resetTenantCache()

	repo := &stubTenantRepo{tenants: map[string]*models.Tenant{
		"tenant-1": {ID: "tenant-1", Domain: strPtr("acme.test")},
	}}

	t.Setenv("TENANT_CTX_SECRET", "secret-current")
	t.Setenv("TENANT_CTX_SECRET_PREV", "")
	t.Setenv("ENV", "prod")

	ts, sig := signedHeaders("tenant-1", "wrong.test", "secret-current", time.Now())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-Id", "tenant-1")
	req.Header.Set("X-Tenant-Domain", "wrong.test")
	req.Header.Set("X-Tenant-Ts", ts)
	req.Header.Set("X-Tenant-Sig", sig)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	TenantContext(repo)(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}
