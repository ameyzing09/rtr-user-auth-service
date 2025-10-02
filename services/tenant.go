package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"rtr-user-auth-service/domain"
	"rtr-user-auth-service/eventbus"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/repositories"
	"rtr-user-auth-service/utils"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type slugConflictError struct {
	suggestions []string
}

func (e slugConflictError) Error() string {
	return domain.ErrTenantSlugTaken.Error()
}

func (e slugConflictError) Suggestions() []string {
	return append([]string(nil), e.suggestions...)
}

var _ TenantService = (*tenantService)(nil)

type tenantService struct {
	db              *gorm.DB
	tenantRepo      TenantRepository
	idempotencyRepo IdempotencyRepository
}

func NewTenantService(db *gorm.DB, tr TenantRepository, idr IdempotencyRepository) *tenantService {
	return &tenantService{
		db:              db,
		tenantRepo:      tr,
		idempotencyRepo: idr,
	}
}

func (s *tenantService) OnboardTenantAsync(ctx context.Context, actor UserRead, req TenantOnboardAsyncRequest, keyHash, requestHash string) (TenantOnboardAsyncResult, bool, error) {
	//priunt all the arguments for debugging
	utils.Debug("[TenantService] OnboardTenantAsync called with actor: %+v, req: %+v, keyHash: %s, requestHash: %s", actor, req, keyHash, requestHash)
	if actor.Role != models.RoleSuperAdmin {
		return TenantOnboardAsyncResult{}, false, domain.ErrSuperadminRequired
	}

	normalizedName := strings.TrimSpace(req.Name)
	//print the normalized name for debugging
	utils.Debug("[TenantService] Normalized tenant name: %s", normalizedName)
	if normalizedName == "" {
		return TenantOnboardAsyncResult{}, false, ErrInvalidInput
	}

	adminName := strings.TrimSpace(req.AdminName)
	//print the admin name for debugging
	utils.Debug("[TenantService] Admin name: %s", adminName)
	if adminName == "" {
		return TenantOnboardAsyncResult{}, false, ErrInvalidInput
	}

	if strings.TrimSpace(req.AdminEmail) == "" {
		return TenantOnboardAsyncResult{}, false, ErrInvalidInput
	}

	adminEmail, err := utils.NormalizeEmail(req.AdminEmail)
	//print the normalized email for debugging
	utils.Debug("[TenantService] Normalized admin email: %s", adminEmail)
	if err != nil {
		return TenantOnboardAsyncResult{}, false, ErrInvalidInput
	}

	var domainPtr *string
	//print the domain if any for debugging
	utils.Debug("[TenantService] Tenant domain: %v", req.Domain)
	if req.Domain != nil {
		normalizedDomain, err := utils.NormalizeDomain(*req.Domain)
		utils.Debug("[TenantService] Normalizing tenant domain: %s", normalizedDomain)
		if err != nil {
			utils.Debug("[TenantService] Error normalizing tenant domain: %v", err)
			return TenantOnboardAsyncResult{}, false, ErrInvalidInput
		}
		domainPtr = &normalizedDomain
	}

	slug, err := utils.NormalizeSlug(normalizedName)
	if err != nil {
		return TenantOnboardAsyncResult{}, false, ErrInvalidInput
	}

	planPtr, err := normalizePlan(req.Plan)
	if err != nil {
		return TenantOnboardAsyncResult{}, false, ErrInvalidInput
	}

	record, err := s.idempotencyRepo.UpsertAndGet(ctx, keyHash, requestHash)
	if err != nil {
		return TenantOnboardAsyncResult{}, false, err
	}

	if record.RequestHash != requestHash {
		return TenantOnboardAsyncResult{}, false, domain.ErrIdempotencyKeyReuseDifferentReq
	}

	if record.Status == models.IdempotencyStatusSuccess && len(record.Response) > 0 {
		var cached TenantOnboardAsyncResult
		if unmarshalErr := json.Unmarshal(record.Response, &cached); unmarshalErr == nil {
			return cached, true, nil
		}
	}

	if domainPtr != nil {
		if existing, err := s.tenantRepo.FindByDomain(ctx, *domainPtr); err == nil && existing != nil {
			return TenantOnboardAsyncResult{}, false, domain.ErrDomainInUse
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return TenantOnboardAsyncResult{}, false, err
		}
	}

	if existing, err := s.tenantRepo.FindBySlug(ctx, slug); err == nil && existing != nil {
		suggestions := utils.SuggestSlugAlternatives(slug)
		return TenantOnboardAsyncResult{}, false, slugConflictError{suggestions: suggestions}
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return TenantOnboardAsyncResult{}, false, err
	}

	var result TenantOnboardAsyncResult
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tenantRepo := repositories.NewGormTenantRepo(tx)
		userRepo := repositories.NewGormUserRepo(tx)
		outboxRepo := repositories.NewGormOutboxRepo(tx)
		bus := eventbus.NewOutboxBus(outboxRepo)

		tenantID := uuid.NewString()
		slugCopy := slug
		tenant := &models.Tenant{
			ID:        tenantID,
			Name:      normalizedName,
			Slug:      &slugCopy,
			Status:    models.TenantPending,
			CreatedBy: &actor.ID,
		}
		if domainPtr != nil {
			tenant.Domain = domainPtr
		}
		if planPtr != nil {
			tenant.Plan = planPtr
		}

		if err := tenantRepo.Create(ctx, tenant); err != nil {
			return err
		}

		tempPassword, err := utils.GenerateTempPassword()
		if err != nil {
			return err
		}
		hashedPassword, err := utils.HashPassword(tempPassword)
		if err != nil {
			return err
		}

		adminUser := &models.User{
			ID:                 uuid.NewString(),
			TenantID:           tenantID,
			Email:              adminEmail,
			Name:               adminName,
			Role:               models.RoleAdmin,
			Password:           hashedPassword,
			IsOwner:            true,
			ForcePasswordReset: true,
		}
		if err := userRepo.Create(ctx, adminUser); err != nil {
			return err
		}

		eventPayload := eventbus.TenantCreatedV1{
			V:             1,
			TenantID:      tenantID,
			Name:          tenant.Name,
			Plan:          string(defaultPlanValue(planPtr)),
			CreatorUserID: actor.ID,
			CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		}
		if domainPtr != nil {
			eventPayload.Domain = *domainPtr
		}
		if err := bus.PublishTenantCreated(ctx, eventPayload); err != nil {
			return err
		}

		result = TenantOnboardAsyncResult{
			TenantID:     tenantID,
			Name:         tenant.Name,
			Domain:       tenant.Domain,
			Slug:         tenant.Slug,
			AdminUserID:  adminUser.ID,
			TempPassword: tempPassword,
			Status:       tenant.Status,
		}
		return nil
	})

	if err != nil {
		var mysqlErr *mysqlDriver.MySQLError
		if errors.As(err, &mysqlErr) {
			switch mysqlErr.Number {
			case 1062:
				if strings.Contains(mysqlErr.Message, "ux_tenants_slug") {
					suggestions := utils.SuggestSlugAlternatives(slug)
					return TenantOnboardAsyncResult{}, false, slugConflictError{suggestions: suggestions}
				}
				if strings.Contains(mysqlErr.Message, "ux_tenants_domain") {
					return TenantOnboardAsyncResult{}, false, domain.ErrDomainInUse
				}
				if strings.Contains(mysqlErr.Message, "ux_users_tenant_email") {
					return TenantOnboardAsyncResult{}, false, domain.ErrEmailInUse
				}
			}
		}
		return TenantOnboardAsyncResult{}, false, err
	}

	responsePayload := map[string]interface{}{
		"tenant": map[string]interface{}{
			"id":   result.TenantID,
			"name": result.Name,
		},
		"admin_user_id": result.AdminUserID,
		"temp_password": result.TempPassword,
		"status":        string(result.Status),
	}
	if result.Domain != nil {
		responsePayload["tenant"].(map[string]interface{})["domain"] = *result.Domain
	}
	if result.Slug != nil {
		responsePayload["tenant"].(map[string]interface{})["slug"] = *result.Slug
	}

	if err := s.idempotencyRepo.SaveResult(ctx, keyHash, models.IdempotencyStatusSuccess, responsePayload); err != nil {
		return result, false, err
	}

	return result, false, nil
}

func (s *tenantService) GetTenant(ctx context.Context, tenantID string) (*models.Tenant, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, ErrInvalidInput
	}
	tenant, err := s.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTenantNotFound
		}
		return nil, err
	}
	return tenant, nil
}

func (s *tenantService) GetTenantStatus(ctx context.Context, tenantID string) (TenantStatusView, error) {
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return TenantStatusView{}, err
	}
	return TenantStatusView{Status: tenant.Status, Steps: []string{}}, nil
}

func (s *tenantService) RetryProvisioning(ctx context.Context, actor UserRead, tenantID string) error {
	if actor.Role != models.RoleSuperAdmin {
		return domain.ErrSuperadminRequired
	}
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		outboxRepo := repositories.NewGormOutboxRepo(tx)
		bus := eventbus.NewOutboxBus(outboxRepo)
		return bus.PublishTenantCreated(ctx, eventbus.TenantCreatedV1{
			V:             1,
			TenantID:      tenant.ID,
			Name:          tenant.Name,
			Plan:          string(defaultPlanValue(tenant.Plan)),
			CreatorUserID: actor.ID,
			CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		})
	})
}

func normalizePlan(plan *models.Plan) (*models.Plan, error) {
	if plan == nil {
		value := models.PlanStarter
		return &value, nil
	}

	switch *plan {
	case models.PlanBasic, models.PlanStarter, models.PlanGrowth, models.PlanEnterprise, models.PlanOnPrem:
		return plan, nil
	default:
		return nil, ErrInvalidInput
	}
}

func defaultPlanValue(plan *models.Plan) models.Plan {
	if plan == nil {
		return models.PlanStarter
	}
	return *plan
}

func (s *tenantService) ListTenants(ctx context.Context, actor UserRead) ([]models.Tenant, error) {
	if actor.Role != models.RoleSuperAdmin {
		return nil, domain.ErrSuperadminRequired
	}

	tenants, err := s.tenantRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	return tenants, nil
}
