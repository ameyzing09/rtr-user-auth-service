package services

import (
	"context"
	"rtr-user-auth-service/models"
	"time"
)

type AuthToken struct {
	Token     string
	ExpiresAt time.Time
}

type UserRead struct {
	ID                 string
	TenantID           string
	Name               string
	Email              string
	Role               models.Role
	Permissions        []string
	ForcePasswordReset bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	LastLogin          *time.Time
}

type LoginInput struct {
	Email    string
	Password string
	TenantID string
}

type CreateUserInput struct {
	Email string
	Name  string
	Role  models.Role
}

type ChangePasswordInput struct {
	CurrentPassword string
	NewPassword     string
}

type SuperadminChangePasswordInput struct {
	UserID   string
	TenantID string
}

type CreateTenantInput struct {
	Name   string
	Domain string
	Email  string
}

type AuthService interface {
	Login(ctx context.Context, input LoginInput) (AuthToken, UserRead, error)
	GetMe(ctx context.Context, userID, tenantID string) (UserRead, error)
	ListUsers(ctx context.Context, tenantID string) ([]UserRead, error)
	CreateUser(ctx context.Context, tenantID string, actor UserRead, input CreateUserInput) (UserRead, string, error)
	ChangePassword(ctx context.Context, tenantID string, actor UserRead, input ChangePasswordInput) error
	SuperadminChangePassword(ctx context.Context, actor UserRead, input SuperadminChangePasswordInput) (string, error)
	// Admin user management (superadmin only)
	AdminListUsers(ctx context.Context, tenantID *string, role *string, search *string, page, limit int) ([]UserRead, int, error)
	AdminGetUser(ctx context.Context, userID string) (UserRead, error)
	AdminResetPassword(ctx context.Context, userID string, newPassword *string, forceChange bool) (string, error)
}
type UserRepository interface {
	EmailExists(ctx context.Context, tenantID, email string) (bool, error)
	Create(ctx context.Context, u *models.User) error
	FindByEmail(ctx context.Context, tenantID, email string) (*models.User, error)
	FindByID(ctx context.Context, tenantID, userID string) (*models.User, error)
	ListByTenant(ctx context.Context, tenantID string) ([]models.User, error)
	UpdatePassword(ctx context.Context, tenantID, userID, hashedPassword string, forcePasswordReset *bool) error
}

type TenantRepository interface {
	Create(ctx context.Context, tenant *models.Tenant) error
	Update(ctx context.Context, tenant *models.Tenant) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, tenantID string) (*models.Tenant, error)
	FindByDomain(ctx context.Context, domain string) (*models.Tenant, error)
	FindBySlug(ctx context.Context, slug string) (*models.Tenant, error)
	ListAll(ctx context.Context) ([]models.Tenant, error)
	ListPaginated(ctx context.Context, page, pageSize int) ([]models.Tenant, int, error)
}

type TenantArchiveRepository interface {
	Create(ctx context.Context, archive *models.TenantArchive) error
	FindByID(ctx context.Context, id string) (*models.TenantArchive, error)
	FindByOriginalTenantID(ctx context.Context, tenantID string) (*models.TenantArchive, error)
	ListPaginated(ctx context.Context, page, pageSize int) ([]models.TenantArchive, int, error)
}

type TenantSettingRepository interface {
	Get(ctx context.Context, tenantID string) (*models.TenantSetting, error)
	PutReplace(ctx context.Context, ts *models.TenantSetting) error
}

type OutboxRepository interface {
	Append(ctx context.Context, aggregateType, aggregateID, eventType string, payload map[string]interface{}) error
}

type IdempotencyRepository interface {
	UpsertAndGet(ctx context.Context, keyHash, requestHash string) (*models.IdempotencyKey, error)
	SaveResult(ctx context.Context, keyHash string, status models.IdempotencyStatus, response map[string]interface{}) error
}

type TenantSettingService interface {
	GetConfiguration(ctx context.Context, tenantID string) (map[string]interface{}, error)
	UpdateConfiguration(ctx context.Context, tenantID string, config map[string]interface{}) (map[string]interface{}, error)
	GetConfigurationValue(ctx context.Context, tenantID, key string) (interface{}, error)
	SetConfigurationValue(ctx context.Context, tenantID, key string, value interface{}) error
	RemoveConfigurationValue(ctx context.Context, tenantID, key string) error
	ResetConfiguration(ctx context.Context, tenantID string) error
}

type TenantOnboardAsyncRequest struct {
	Name       string
	Domain     *string
	AdminName  string
	AdminEmail string
	Plan       *models.Plan
	IsTrial    bool
}

type TenantOnboardAsyncResult struct {
	TenantID     string
	Name         string
	Domain       *string
	Slug         *string
	AdminUserID  string
	TempPassword string
	Status       models.TenantStatus
}

type TenantStatusView struct {
	Status models.TenantStatus
	Steps  []string
}

type CreateTenantReq struct {
	Name    string
	Domain  *string
	Plan    models.Plan
	IsTrial bool
}

type UpdateTenantReq struct {
	Name   *string
	Domain *string
	Plan   *models.Plan
	Status *models.TenantStatus
}

type TenantDTO struct {
	ID           string
	Name         string
	Domain       *string
	Slug         *string
	Plan         *models.Plan
	Status       models.TenantStatus
	CreatedBy    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	FailedReason *string
}

type TenantListResult struct {
	Tenants  []TenantDTO
	Total    int
	Page     int
	PageSize int
}

type TenantArchiveDTO struct {
	ID           string
	Name         string
	Domain       *string
	Slug         *string
	Plan         *models.Plan
	Status       models.TenantStatus
	CreatedBy    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	FailedReason *string
	DeletedBy    string
	DeletedAt    time.Time
	DeleteReason *string
}

type TenantArchiveListResult struct {
	Archives []TenantArchiveDTO
	Total    int
	Page     int
	PageSize int
}

type TenantService interface {
	OnboardTenantAsync(ctx context.Context, actor UserRead, req TenantOnboardAsyncRequest, keyHash, requestHash string) (TenantOnboardAsyncResult, bool, error)
	CreateTenant(ctx context.Context, req CreateTenantReq, actorID string) (TenantDTO, error)
	GetTenant(ctx context.Context, id string) (TenantDTO, error)
	UpdateTenant(ctx context.Context, id string, req UpdateTenantReq, actorID string) (TenantDTO, error)
	DeleteTenant(ctx context.Context, id string, actorID string) error
	ListTenants(ctx context.Context, page, pageSize int) (TenantListResult, error)
	GetTenantStatus(ctx context.Context, tenantID string) (TenantStatusView, error)
	RetryProvisioning(ctx context.Context, actor UserRead, tenantID string) error

	// Archive management
	ListArchivedTenants(ctx context.Context, page, pageSize int) (TenantArchiveListResult, error)
	GetArchivedTenant(ctx context.Context, id string) (TenantArchiveDTO, error)
}
