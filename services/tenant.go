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
	db                *gorm.DB
	tenantRepo        TenantRepository
	tenantArchiveRepo TenantArchiveRepository
	idempotencyRepo   IdempotencyRepository
	subscriptionSvc   SubscriptionService
}

func NewTenantService(db *gorm.DB, tr TenantRepository, tar TenantArchiveRepository, idr IdempotencyRepository, subSvc SubscriptionService) *tenantService {
	return &tenantService{
		db:                db,
		tenantRepo:        tr,
		tenantArchiveRepo: tar,
		idempotencyRepo:   idr,
		subscriptionSvc:   subSvc,
	}
}

func (s *tenantService) OnboardTenantAsync(ctx context.Context, actor UserRead, req TenantOnboardAsyncRequest, keyHash, requestHash string) (TenantOnboardAsyncResult, bool, error) {
	//priunt all the arguments for debugging
	utils.Debug("[TenantService] OnboardTenantAsync called with actorID: %s, actorRole: %s", actor.ID, actor.Role)
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

		tenantID := uuid.NewString()
		slugCopy := slug
		tenant := &models.Tenant{
			ID:        tenantID,
			Name:      normalizedName,
			Slug:      &slugCopy,
			Status:    models.TenantActive,
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

		// Create subscription
		subscriptionRepo := repositories.NewSubscriptionRepository(tx)
		subscriptionSvc := NewSubscriptionService(subscriptionRepo)
		_, err := subscriptionSvc.CreateSubscription(ctx, tenantID, defaultPlanValue(planPtr), req.IsTrial, actor.ID)
		if err != nil {
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

		// Skipping event publishing - tenant is immediately active
		// No async onboarding queue processing needed

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

func (s *tenantService) CreateTenant(ctx context.Context, req CreateTenantReq, actorID string) (TenantDTO, error) {
	normalizedName := strings.TrimSpace(req.Name)
	if normalizedName == "" {
		return TenantDTO{}, ErrInvalidInput
	}

	var domainPtr *string
	if req.Domain != nil {
		normalizedDomain, err := utils.NormalizeDomain(*req.Domain)
		if err != nil {
			return TenantDTO{}, ErrInvalidInput
		}
		domainPtr = &normalizedDomain
	}

	slug, err := utils.NormalizeSlug(normalizedName)
	if err != nil {
		return TenantDTO{}, ErrInvalidInput
	}

	// Check for domain conflicts
	if domainPtr != nil {
		if existing, err := s.tenantRepo.FindByDomain(ctx, *domainPtr); err == nil && existing != nil {
			return TenantDTO{}, domain.ErrDomainInUse
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return TenantDTO{}, err
		}
	}

	// Check for slug conflicts
	if existing, err := s.tenantRepo.FindBySlug(ctx, slug); err == nil && existing != nil {
		return TenantDTO{}, domain.ErrTenantSlugTaken
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return TenantDTO{}, err
	}

	var result TenantDTO
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tenantRepo := repositories.NewGormTenantRepo(tx)
		subscriptionRepo := repositories.NewSubscriptionRepository(tx)
		subscriptionSvc := NewSubscriptionService(subscriptionRepo)

		tenantID := uuid.NewString()
		slugCopy := slug
		tenant := &models.Tenant{
			ID:        tenantID,
			Name:      normalizedName,
			Slug:      &slugCopy,
			Status:    models.TenantActive,
			CreatedBy: &actorID,
			Plan:      &req.Plan,
		}
		if domainPtr != nil {
			tenant.Domain = domainPtr
		}

		if err := tenantRepo.Create(ctx, tenant); err != nil {
			return err
		}

		// Create subscription
		_, err := subscriptionSvc.CreateSubscription(ctx, tenantID, req.Plan, req.IsTrial, actorID)
		if err != nil {
			return err
		}

		result = TenantDTO{
			ID:        tenant.ID,
			Name:      tenant.Name,
			Domain:    tenant.Domain,
			Slug:      tenant.Slug,
			Plan:      tenant.Plan,
			Status:    tenant.Status,
			CreatedBy: tenant.CreatedBy,
			CreatedAt: tenant.CreatedAt,
			UpdatedAt: tenant.UpdatedAt,
		}

		return nil
	})

	if err != nil {
		var mysqlErr *mysqlDriver.MySQLError
		if errors.As(err, &mysqlErr) {
			switch mysqlErr.Number {
			case 1062:
				if strings.Contains(mysqlErr.Message, "ux_tenants_slug") {
					return TenantDTO{}, domain.ErrTenantSlugTaken
				}
				if strings.Contains(mysqlErr.Message, "ux_tenants_domain") {
					return TenantDTO{}, domain.ErrDomainInUse
				}
			}
		}
		return TenantDTO{}, err
	}

	return result, nil
}

func (s *tenantService) GetTenant(ctx context.Context, id string) (TenantDTO, error) {
	if strings.TrimSpace(id) == "" {
		return TenantDTO{}, ErrInvalidInput
	}

	tenant, err := s.tenantRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TenantDTO{}, domain.ErrTenantNotFound
		}
		return TenantDTO{}, err
	}

	return TenantDTO{
		ID:           tenant.ID,
		Name:         tenant.Name,
		Domain:       tenant.Domain,
		Slug:         tenant.Slug,
		Plan:         tenant.Plan,
		Status:       tenant.Status,
		CreatedBy:    tenant.CreatedBy,
		CreatedAt:    tenant.CreatedAt,
		UpdatedAt:    tenant.UpdatedAt,
		FailedReason: tenant.FailedReason,
	}, nil
}

func (s *tenantService) UpdateTenant(ctx context.Context, id string, req UpdateTenantReq, actorID string) (TenantDTO, error) {
	if strings.TrimSpace(id) == "" {
		return TenantDTO{}, ErrInvalidInput
	}

	tenant, err := s.tenantRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TenantDTO{}, domain.ErrTenantNotFound
		}
		return TenantDTO{}, err
	}

	// Update fields if provided
	if req.Name != nil {
		normalizedName := strings.TrimSpace(*req.Name)
		if normalizedName == "" {
			return TenantDTO{}, ErrInvalidInput
		}
		tenant.Name = normalizedName
	}

	if req.Domain != nil {
		if *req.Domain == "" {
			tenant.Domain = nil
		} else {
			normalizedDomain, err := utils.NormalizeDomain(*req.Domain)
			if err != nil {
				return TenantDTO{}, ErrInvalidInput
			}
			tenant.Domain = &normalizedDomain
		}
	}

	if req.Plan != nil {
		tenant.Plan = req.Plan
	}

	if req.Status != nil {
		tenant.Status = *req.Status
	}

	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return TenantDTO{}, err
	}

	return TenantDTO{
		ID:           tenant.ID,
		Name:         tenant.Name,
		Domain:       tenant.Domain,
		Slug:         tenant.Slug,
		Plan:         tenant.Plan,
		Status:       tenant.Status,
		CreatedBy:    tenant.CreatedBy,
		CreatedAt:    tenant.CreatedAt,
		UpdatedAt:    tenant.UpdatedAt,
		FailedReason: tenant.FailedReason,
	}, nil
}

func (s *tenantService) DeleteTenant(ctx context.Context, id string, actorID string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidInput
	}

	// Check if tenant exists
	tenant, err := s.tenantRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrTenantNotFound
		}
		return err
	}

	// Use transaction for archiving and soft delete
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create archive record
		archive := &models.TenantArchive{}
		deleteReason := "Tenant deleted by admin"
		archive.FromTenant(tenant, actorID, &deleteReason)

		if err := tx.Create(archive).Error; err != nil {
			return err
		}

		// Update tenant status to DELETED
		if err := tx.Model(tenant).Update("status", models.TenantDeleted).Error; err != nil {
			return err
		}

		// Soft delete the tenant (sets deleted_at)
		if err := tx.Delete(tenant).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *tenantService) ListTenants(ctx context.Context, page, pageSize int) (TenantListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tenants, total, err := s.tenantRepo.ListPaginated(ctx, page, pageSize)
	if err != nil {
		return TenantListResult{}, err
	}

	utils.Debug("[TenantService] ListTenants found %d tenants (total: %d, page: %d, pageSize: %d)", len(tenants), total, page, pageSize)

	tenantDTOs := make([]TenantDTO, len(tenants))
	for i, tenant := range tenants {
		tenantDTOs[i] = TenantDTO{
			ID:           tenant.ID,
			Name:         tenant.Name,
			Domain:       tenant.Domain,
			Slug:         tenant.Slug,
			Plan:         tenant.Plan,
			Status:       tenant.Status,
			CreatedBy:    tenant.CreatedBy,
			CreatedAt:    tenant.CreatedAt,
			UpdatedAt:    tenant.UpdatedAt,
			FailedReason: tenant.FailedReason,
		}
	}

	return TenantListResult{
		Tenants:  tenantDTOs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ListArchivedTenants returns paginated list of archived tenants
func (s *tenantService) ListArchivedTenants(ctx context.Context, page, pageSize int) (TenantArchiveListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	archives, total, err := s.tenantArchiveRepo.ListPaginated(ctx, page, pageSize)
	if err != nil {
		return TenantArchiveListResult{}, err
	}

	archiveDTOs := make([]TenantArchiveDTO, len(archives))
	for i, archive := range archives {
		archiveDTOs[i] = TenantArchiveDTO{
			ID:           archive.ID,
			Name:         archive.Name,
			Domain:       archive.Domain,
			Slug:         archive.Slug,
			Plan:         archive.Plan,
			Status:       archive.Status,
			CreatedBy:    archive.CreatedBy,
			CreatedAt:    archive.CreatedAt,
			UpdatedAt:    archive.UpdatedAt,
			FailedReason: archive.FailedReason,
			DeletedBy:    archive.DeletedBy,
			DeletedAt:    archive.DeletedAt,
			DeleteReason: archive.DeleteReason,
		}
	}

	return TenantArchiveListResult{
		Archives: archiveDTOs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetArchivedTenant returns archived tenant by ID
func (s *tenantService) GetArchivedTenant(ctx context.Context, id string) (TenantArchiveDTO, error) {
	if strings.TrimSpace(id) == "" {
		return TenantArchiveDTO{}, ErrInvalidInput
	}

	archive, err := s.tenantArchiveRepo.FindByOriginalTenantID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TenantArchiveDTO{}, domain.ErrTenantNotFound
		}
		return TenantArchiveDTO{}, err
	}

	return TenantArchiveDTO{
		ID:           archive.ID,
		Name:         archive.Name,
		Domain:       archive.Domain,
		Slug:         archive.Slug,
		Plan:         archive.Plan,
		Status:       archive.Status,
		CreatedBy:    archive.CreatedBy,
		CreatedAt:    archive.CreatedAt,
		UpdatedAt:    archive.UpdatedAt,
		FailedReason: archive.FailedReason,
		DeletedBy:    archive.DeletedBy,
		DeletedAt:    archive.DeletedAt,
		DeleteReason: archive.DeleteReason,
	}, nil
}
