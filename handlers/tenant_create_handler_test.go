package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	errcodes "rtr-user-auth-service/errors"
	"rtr-user-auth-service/middleware"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/repositories"
	"rtr-user-auth-service/services"
	"rtr-user-auth-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type tenantTestEnv struct {
	t          *testing.T
	db         *gorm.DB
	router     *gin.Engine
	tenantRepo *repositories.GormTenantRepo
}

func newTenantTestEnv(t *testing.T) tenantTestEnv {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Tenant{}, &models.User{}, &models.Outbox{}, &models.IdempotencyKey{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	tenantRepo := repositories.NewGormTenantRepo(db)
	idempotencyRepo := repositories.NewGormIdempotencyRepo(db)
	tenantService := services.NewTenantService(db, tenantRepo, idempotencyRepo)
	handler := NewTenantCreateHandler(tenantService)

	router := gin.New()
	router.Use(gin.Recovery())
	admin := router.Group("/")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.POST("/tenant/create", handler.Create)
		admin.GET("/tenant/:id", handler.Get)
	}

	return tenantTestEnv{t: t, db: db, router: router, tenantRepo: tenantRepo}
}

func (env tenantTestEnv) doJSON(method, path string, body interface{}, token string, headers map[string]string) *httptest.ResponseRecorder {
	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			env.t.Fatalf("failed to marshal payload: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	return w
}

func makeToken(t *testing.T, role models.Role) string {
	claims := &utils.Claims{
		UserID:   "user-super",
		TenantID: "system",
		Email:    "super@example.com",
		Role:     string(role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func strp(v string) *string {
	return &v
}

func TestTenantCreateHandler_SuperAdminFlow(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	env := newTenantTestEnv(t)

	plan := "STARTER"
	req := TenantCreateRequest{
		Name:       "Acme Corporation",
		Domain:     strp("acme.com"),
		AdminName:  "Alice Johnson",
		AdminEmail: "alice@acme.com",
		Plan:       &plan,
	}

	token := makeToken(t, models.RoleSuperAdmin)
	headers := map[string]string{"Idempotency-Key": "key-001"}

	w := env.doJSON(http.MethodPost, "/tenant/create", req, token, headers)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	var resp TenantCreateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Tenant.ID == "" {
		t.Fatalf("expected tenant id in response")
	}
	if resp.TempPassword == "" {
		t.Fatalf("expected temp password")
	}
	if resp.Tenant.Slug == nil || *resp.Tenant.Slug == "" {
		t.Fatalf("expected slug to be set")
	}

	var tenant models.Tenant
	if err := env.db.First(&tenant, "id = ?", resp.Tenant.ID).Error; err != nil {
		t.Fatalf("expected tenant persisted: %v", err)
	}
	if tenant.Status != models.TenantPending {
		t.Fatalf("expected tenant status pending, got %s", tenant.Status)
	}
	if tenant.Plan == nil || *tenant.Plan != models.PlanStarter {
		t.Fatalf("expected plan starter, got %v", tenant.Plan)
	}
	if tenant.CreatedBy == nil || *tenant.CreatedBy != "user-super" {
		t.Fatalf("expected created_by user-super, got %v", tenant.CreatedBy)
	}

	var adminUser models.User
	if err := env.db.First(&adminUser, "tenant_id = ?", resp.Tenant.ID).Error; err != nil {
		t.Fatalf("expected admin user persisted: %v", err)
	}
	if !adminUser.IsOwner {
		t.Fatalf("expected admin user to be owner")
	}
	if !adminUser.ForcePasswordReset {
		t.Fatalf("expected force password change true")
	}

	keyHash := utils.HashKey("key-001")
	var idem models.IdempotencyKey
	if err := env.db.First(&idem, "key_hash = ?", keyHash).Error; err != nil {
		t.Fatalf("expected idempotency key persisted: %v", err)
	}
	if idem.Status != models.IdempotencyStatusSuccess {
		t.Fatalf("expected idempotency status success, got %s", idem.Status)
	}
	if len(idem.Response) == 0 {
		t.Fatalf("expected stored idempotency response")
	}

	var outbox models.Outbox
	if err := env.db.First(&outbox).Error; err != nil {
		t.Fatalf("expected outbox row: %v", err)
	}
	if outbox.Type != "tenant.created" {
		t.Fatalf("expected outbox type tenant.created, got %s", outbox.Type)
	}

	repeat := env.doJSON(http.MethodPost, "/tenant/create", req, token, headers)
	if repeat.Code != http.StatusOK {
		t.Fatalf("expected cached 200, got %d: %s", repeat.Code, repeat.Body.String())
	}
	var cached TenantCreateResponse
	if err := json.Unmarshal(repeat.Body.Bytes(), &cached); err != nil {
		t.Fatalf("failed to unmarshal cached response: %v", err)
	}
	if cached.Tenant.ID != resp.Tenant.ID {
		t.Fatalf("expected cached response to match initial")
	}

	diffReq := req
	diffReq.AdminEmail = "other@acme.com"
	diff := env.doJSON(http.MethodPost, "/tenant/create", diffReq, token, headers)
	if diff.Code != http.StatusConflict {
		t.Fatalf("expected 409 on differing idempotent request, got %d", diff.Code)
	}
	var diffBody map[string]interface{}
	if err := json.Unmarshal(diff.Body.Bytes(), &diffBody); err != nil {
		t.Fatalf("failed to decode diff response: %v", err)
	}
	if diffBody["code"] != errcodes.ErrCodeIdempotencyKeyReuseDiff {
		t.Fatalf("expected code %s, got %v", errcodes.ErrCodeIdempotencyKeyReuseDiff, diffBody["code"])
	}

	getResp := env.doJSON(http.MethodGet, "/tenant/"+resp.Tenant.ID, nil, token, nil)
	if getResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for tenant get, got %d: %s", getResp.Code, getResp.Body.String())
	}
	var tenantView TenantGetResponse
	if err := json.Unmarshal(getResp.Body.Bytes(), &tenantView); err != nil {
		t.Fatalf("failed to decode tenant get response: %v", err)
	}
	if tenantView.ID != resp.Tenant.ID || tenantView.Status != string(models.TenantPending) {
		t.Fatalf("unexpected tenant view: %+v", tenantView)
	}
}

func TestTenantCreateHandler_NonSuperadminForbidden(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	env := newTenantTestEnv(t)

	token := makeToken(t, models.RoleAdmin)
	req := TenantCreateRequest{
		Name:       "Acme",
		AdminName:  "Alice",
		AdminEmail: "alice@acme.com",
	}

	w := env.doJSON(http.MethodPost, "/tenant/create", req, token, map[string]string{"Idempotency-Key": "key-002"})
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestTenantCreateHandler_SlugConflictProvidesSuggestions(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	env := newTenantTestEnv(t)

	slug := "acme-hq"
	tenant := &models.Tenant{
		ID:     "existing-tenant",
		Name:   "Acme HQ",
		Slug:   &slug,
		Status: models.TenantActive,
	}
	if err := env.db.Create(tenant).Error; err != nil {
		t.Fatalf("failed to seed tenant: %v", err)
	}

	req := TenantCreateRequest{
		Name:       "Acme HQ",
		AdminName:  "Alice",
		AdminEmail: "alice@acme.com",
	}

	token := makeToken(t, models.RoleSuperAdmin)
	w := env.doJSON(http.MethodPost, "/tenant/create", req, token, map[string]string{"Idempotency-Key": "key-003"})
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
	var body struct {
		Code        string   `json:"code"`
		Suggestions []string `json:"suggestions"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Code != errcodes.ErrCodeTenantSlugTaken {
		t.Fatalf("expected code %s, got %s", errcodes.ErrCodeTenantSlugTaken, body.Code)
	}
	if len(body.Suggestions) == 0 {
		t.Fatalf("expected suggestions in response")
	}
}
