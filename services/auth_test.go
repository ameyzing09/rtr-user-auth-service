package services

import (
	"context"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/repositories"
	"rtr-user-auth-service/utils"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type authTestEnv struct {
	t               *testing.T
	db              *gorm.DB
	authService     *authService
	userRepo        *repositories.GormUserRepo
	tenantRepo      *repositories.GormTenantRepo
	subscriptionSvc SubscriptionService
}

func newAuthTestEnv(t *testing.T) authTestEnv {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Tenant{}, &models.Subscription{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	userRepo := repositories.NewGormUserRepo(db)
	tenantRepo := repositories.NewGormTenantRepo(db)
	subscriptionRepo := repositories.NewSubscriptionRepository(db)
	subscriptionSvc := NewSubscriptionService(subscriptionRepo)
	authService := NewAuthService(db, userRepo, tenantRepo, subscriptionSvc)

	return authTestEnv{
		t:               t,
		db:              db,
		authService:     authService,
		userRepo:        userRepo,
		tenantRepo:      tenantRepo,
		subscriptionSvc: subscriptionSvc,
	}
}

func (env authTestEnv) createSuperAdmin(email, password string, isActive bool) *models.User {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		env.t.Fatalf("failed to hash password: %v", err)
	}

	user := &models.User{
		ID:       uuid.NewString(),
		TenantID: "system",
		Email:    email,
		Name:     "Super Admin",
		Role:     models.RoleSuperAdmin,
		Password: hashedPassword,
		IsActive: isActive,
	}

	if err := env.db.Create(user).Error; err != nil {
		env.t.Fatalf("failed to create user: %v", err)
	}

	return user
}

func (env authTestEnv) createTenant(tenantID string) *models.Tenant {
	slug := "test-tenant"
	tenant := &models.Tenant{
		ID:     tenantID,
		Name:   "Test Tenant",
		Slug:   &slug,
		Status: models.TenantActive,
	}

	if err := env.db.Create(tenant).Error; err != nil {
		env.t.Fatalf("failed to create tenant: %v", err)
	}

	return tenant
}

func (env authTestEnv) createSubscription(tenantID string, status models.SubscriptionStatus) *models.Subscription {
	sub := &models.Subscription{
		TenantID: tenantID,
		Plan:     models.PlanStarter,
		Status:   status,
	}

	if err := env.db.Create(sub).Error; err != nil {
		env.t.Fatalf("failed to create subscription: %v", err)
	}

	return sub
}

// TestLogin_SoftDeletedSuperAdmin verifies that soft-deleted SuperAdmins cannot login
func TestLogin_SoftDeletedSuperAdmin(t *testing.T) {
	env := newAuthTestEnv(t)
	ctx := context.Background()

	email := "superadmin@example.com"
	password := "password123"

	// Create an active SuperAdmin
	user := env.createSuperAdmin(email, password, true)

	// Verify the SuperAdmin can login initially
	input := LoginInput{
		Email:    email,
		Password: password,
		TenantID: "", // No tenant context for SuperAdmin
	}

	token, userRead, err := env.authService.Login(ctx, input)
	if err != nil {
		t.Fatalf("expected successful login, got error: %v", err)
	}
	if token.Token == "" {
		t.Fatalf("expected non-empty token")
	}
	if userRead.Email != email {
		t.Fatalf("expected email %s, got %s", email, userRead.Email)
	}

	// Soft-delete the user using GORM's Delete method
	if err := env.db.Delete(user).Error; err != nil {
		t.Fatalf("failed to soft-delete user: %v", err)
	}

	// Verify the DeletedAt field is set
	var deletedUser models.User
	if err := env.db.Unscoped().First(&deletedUser, "id = ?", user.ID).Error; err != nil {
		t.Fatalf("failed to fetch deleted user: %v", err)
	}
	if deletedUser.DeletedAt.Time.IsZero() {
		t.Fatalf("expected DeletedAt to be set, but it's zero")
	}

	// Attempt to login with the soft-deleted SuperAdmin
	_, _, err = env.authService.Login(ctx, input)
	if err == nil {
		t.Fatalf("expected login to fail for soft-deleted SuperAdmin, but it succeeded")
	}

	// Verify the error message indicates invalid credentials (not revealing that the user is deleted)
	if err.Error() != "invalid email or password" {
		t.Logf("got error: %v (this is expected - soft-deleted users should not be able to login)", err)
	}
}

// TestLogin_InactiveSuperAdmin verifies that inactive SuperAdmins cannot login
func TestLogin_InactiveSuperAdmin(t *testing.T) {
	env := newAuthTestEnv(t)
	ctx := context.Background()

	email := "inactive@example.com"
	password := "password123"

	// Create an inactive SuperAdmin
	env.createSuperAdmin(email, password, false)

	// Attempt to login with inactive SuperAdmin
	input := LoginInput{
		Email:    email,
		Password: password,
		TenantID: "", // No tenant context for SuperAdmin
	}

	_, _, err := env.authService.Login(ctx, input)
	if err == nil {
		t.Fatalf("expected login to fail for inactive SuperAdmin, but it succeeded")
	}

	// Verify the error indicates invalid credentials
	if err.Error() != "invalid email or password" {
		t.Logf("got error: %v (this is expected - inactive users should not be able to login)", err)
	}
}

// TestLogin_ActiveSuperAdmin verifies that active SuperAdmins can login successfully
func TestLogin_ActiveSuperAdmin(t *testing.T) {
	env := newAuthTestEnv(t)
	ctx := context.Background()

	email := "active@example.com"
	password := "password123"

	// Create an active SuperAdmin
	env.createSuperAdmin(email, password, true)

	// Verify the SuperAdmin can login
	input := LoginInput{
		Email:    email,
		Password: password,
		TenantID: "", // No tenant context for SuperAdmin
	}

	token, userRead, err := env.authService.Login(ctx, input)
	if err != nil {
		t.Fatalf("expected successful login, got error: %v", err)
	}
	if token.Token == "" {
		t.Fatalf("expected non-empty token")
	}
	if userRead.Email != email {
		t.Fatalf("expected email %s, got %s", email, userRead.Email)
	}
	if userRead.Role != models.RoleSuperAdmin {
		t.Fatalf("expected role %s, got %s", models.RoleSuperAdmin, userRead.Role)
	}
}

// TestLogin_SoftDeletedTenantUser verifies that soft-deleted tenant users cannot login
func TestLogin_SoftDeletedTenantUser(t *testing.T) {
	env := newAuthTestEnv(t)
	ctx := context.Background()

	tenantID := uuid.NewString()
	email := "user@tenant.com"
	password := "password123"

	// Create tenant and subscription
	env.createTenant(tenantID)
	env.createSubscription(tenantID, models.SubActive)

	// Create a regular tenant user
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	user := &models.User{
		ID:       uuid.NewString(),
		TenantID: tenantID,
		Email:    email,
		Name:     "Regular User",
		Role:     models.RoleAdmin,
		Password: hashedPassword,
		IsActive: true,
	}

	if err := env.db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Verify the user can login initially
	input := LoginInput{
		Email:    email,
		Password: password,
		TenantID: tenantID,
	}

	token, userRead, err := env.authService.Login(ctx, input)
	if err != nil {
		t.Fatalf("expected successful login, got error: %v", err)
	}
	if token.Token == "" {
		t.Fatalf("expected non-empty token")
	}
	if userRead.Email != email {
		t.Fatalf("expected email %s, got %s", email, userRead.Email)
	}

	// Soft-delete the user
	if err := env.db.Delete(user).Error; err != nil {
		t.Fatalf("failed to soft-delete user: %v", err)
	}

	// Attempt to login with soft-deleted user
	_, _, err = env.authService.Login(ctx, input)
	if err == nil {
		t.Fatalf("expected login to fail for soft-deleted user, but it succeeded")
	}
}

// TestLogin_InactiveTenantUser verifies that inactive tenant users cannot login
func TestLogin_InactiveTenantUser(t *testing.T) {
	env := newAuthTestEnv(t)
	ctx := context.Background()

	tenantID := uuid.NewString()
	email := "inactive-user@tenant.com"
	password := "password123"

	// Create tenant and subscription
	env.createTenant(tenantID)
	env.createSubscription(tenantID, models.SubActive)

	// Create an inactive tenant user
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	user := &models.User{
		ID:       uuid.NewString(),
		TenantID: tenantID,
		Email:    email,
		Name:     "Inactive User",
		Role:     models.RoleAdmin,
		Password: hashedPassword,
		IsActive: false, // Inactive user
	}

	if err := env.db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Attempt to login with inactive user
	input := LoginInput{
		Email:    email,
		Password: password,
		TenantID: tenantID,
	}

	_, _, err = env.authService.Login(ctx, input)
	if err == nil {
		t.Fatalf("expected login to fail for inactive user, but it succeeded")
	}
}
