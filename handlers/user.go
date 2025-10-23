package handlers

import (
	"net/http"
	"strings"
	"time"

	"rtr-user-auth-service/config"
	"rtr-user-auth-service/domain"
	errcodes "rtr-user-auth-service/errors"
	"rtr-user-auth-service/middleware"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/services"
	"rtr-user-auth-service/utils"
	"errors"
	"rtr-user-auth-service/utils/httpx"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	authService          services.AuthService
	tenantSettingService services.TenantSettingService
}

func NewUserHandler(authService services.AuthService, tenantSettingService services.TenantSettingService) *UserHandler {
	return &UserHandler{
		authService:          authService,
		tenantSettingService: tenantSettingService,
	}
}

func (h *UserHandler) Login(c *gin.Context) {
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	// Extract tenant ID from context (set by TenantContext middleware)
	tenantID := middleware.GetTenantIDFromContext(c)

	token, user, err := h.authService.Login(c, services.LoginInput{
		Email:    loginReq.Email,
		Password: loginReq.Password,
		TenantID: tenantID,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	// Audit: Successful login
	if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
		clientIP := middleware.GetClientIP(c)
		userAgent := middleware.GetUserAgent(c)
		actorRoleStr := string(user.Role)
		_ = auditSvc.Log(c.Request.Context(), services.AuditLogEntry{
			Action:        utils.AuditActionLoginSuccess,
			ActorID:       &user.ID,
			ActorTenantID: &user.TenantID,
			ActorRole:     &actorRoleStr,
			Status:        models.AuditStatusSuccess,
			IPAddress:     utils.StringPtr(clientIP),
			UserAgent:     utils.StringPtr(userAgent),
			Metadata: map[string]interface{}{
				"email": user.Email,
			},
		})
	}

	c.Header("X-Tenant-ID", user.TenantID)

	// Set httpOnly cookie for JWT token
	setCookies(c, token.Token)

	// Fetch tenant branding if user is not a SuperAdmin
	var tenantBranding *TenantBranding
	if user.Role != models.RoleSuperAdmin && user.TenantID != "" {
		branding := h.resolveTenantBranding(c, user.TenantID)
		if branding != nil {
			tenantBranding = branding
		}
	}

	response := LoginResponse{
		Token:          token.Token,
		ExpiresAt:      token.ExpiresAt,
		User:           user,
		TenantBranding: tenantBranding,
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) AdminLogin(c *gin.Context) {
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	// AdminLogin has no tenant context - TenantID will be empty
	// The Login service will handle SuperAdmin authentication without tenant isolation
	token, user, err := h.authService.Login(c, services.LoginInput{
		Email:    loginReq.Email,
		Password: loginReq.Password,
		TenantID: "", // No tenant context for admin login
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	if user.Role != models.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    errcodes.ErrCodeSuperadminRequired,
			"message": domain.ErrSuperadminRequired.Error(),
		})
		return
	}

	// Audit: Successful admin login
	if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
		clientIP := middleware.GetClientIP(c)
		userAgent := middleware.GetUserAgent(c)
		actorRoleStr := string(user.Role)
		_ = auditSvc.Log(c.Request.Context(), services.AuditLogEntry{
			Action:        utils.AuditActionLoginSuccess,
			ActorID:       &user.ID,
			ActorTenantID: &user.TenantID,
			ActorRole:     &actorRoleStr,
			Status:        models.AuditStatusSuccess,
			IPAddress:     utils.StringPtr(clientIP),
			UserAgent:     utils.StringPtr(userAgent),
			Metadata: map[string]interface{}{
				"email":      user.Email,
				"admin_login": true,
			},
		})
	}

	branding := resolvePlatformBranding(config.Get())

	// Set httpOnly cookie for JWT token
	setCookies(c, token.Token)

	response := LoginResponse{
		Token:            token.Token,
		ExpiresAt:        token.ExpiresAt,
		User:             user,
		PlatformBranding: &branding,
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) GetMe(c *gin.Context) {
	actor := c.MustGet("actor").(services.UserRead)
	user, err := h.authService.GetMe(c, actor.ID, actor.TenantID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	actor := c.MustGet("actor").(services.UserRead)
	users, err := h.authService.ListUsers(c, actor.TenantID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	actor := c.MustGet("actor").(services.UserRead)

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	user, tempPassword, err := h.authService.CreateUser(c, actor.TenantID, actor, services.CreateUserInput{
		Email: req.Email,
		Name:  req.Name,
		Role:  req.Role,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":               user,
		"temporary_password": tempPassword,
	})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	actor := c.MustGet("actor").(services.UserRead)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	if err := h.authService.ChangePassword(c, actor.TenantID, actor, services.ChangePasswordInput{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}); err != nil {
		httpx.HandleError(c, err)
		return
	}

	// Audit: Password change
	if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
		clientIP := middleware.GetClientIP(c)
		userAgent := middleware.GetUserAgent(c)
		actorRoleStr := string(actor.Role)
		_ = auditSvc.Log(c.Request.Context(), services.AuditLogEntry{
			Action:        utils.AuditActionPasswordChange,
			ActorID:       &actor.ID,
			ActorTenantID: &actor.TenantID,
			ActorRole:     &actorRoleStr,
			Status:        models.AuditStatusSuccess,
			IPAddress:     utils.StringPtr(clientIP),
			UserAgent:     utils.StringPtr(userAgent),
		})
	}

	dropClientCache(c)
	c.Status(http.StatusNoContent)
}

func (h *UserHandler) SuperadminChangePassword(c *gin.Context) {
	actor := c.MustGet("actor").(services.UserRead)

	var req SuperadminChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	tempPassword, err := h.authService.SuperadminChangePassword(c, actor, services.SuperadminChangePasswordInput{
		UserID:   req.UserID,
		TenantID: req.TenantID,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	response := SuperadminChangePasswordResponse{
		TemporaryPassword: tempPassword,
	}

	dropClientCache(c)
	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) Logout(c *gin.Context) {
	// Audit: Logout
	if actorValue, exists := c.Get("actor"); exists {
		if actor, ok := actorValue.(services.UserRead); ok {
			if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
				clientIP := middleware.GetClientIP(c)
				userAgent := middleware.GetUserAgent(c)
				actorRoleStr := string(actor.Role)
				_ = auditSvc.Log(c.Request.Context(), services.AuditLogEntry{
					Action:        utils.AuditActionLogout,
					ActorID:       &actor.ID,
					ActorTenantID: &actor.TenantID,
					ActorRole:     &actorRoleStr,
					Status:        models.AuditStatusSuccess,
					IPAddress:     utils.StringPtr(clientIP),
					UserAgent:     utils.StringPtr(userAgent),
				})
			}
		}
	}

	// Clear cookies
	clearCookies(c)
	dropClientCache(c)
	c.Status(http.StatusNoContent)
}

// AdminListUsers lists all users across all tenants or within a specific tenant
// Requires superadmin role and SYS_USER_LIST permission
func (h *UserHandler) AdminListUsers(c *gin.Context) {
	var req AdminListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 50
	}

	// Get all users (this will be implemented in the service)
	users, total, err := h.authService.AdminListUsers(c.Request.Context(), req.TenantID, req.Role, req.Search, req.Page, req.Limit)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	// Transform to response DTOs
	response := AdminListUsersResponse{
		Users: make([]AdminUserDetail, 0, len(users)),
		Total: total,
		Page:  req.Page,
		Limit: req.Limit,
	}

	for _, user := range users {
		response.Users = append(response.Users, AdminUserDetail{
			ID:                 user.ID,
			TenantID:           user.TenantID,
			Name:               user.Name,
			Email:              user.Email,
			Role:               string(user.Role),
			ForcePasswordReset: user.ForcePasswordReset,
			CreatedAt:          user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          user.UpdatedAt.Format(time.RFC3339),
			LastLogin:          nil,
		})
	}

	// Audit log
	if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
		actor := c.MustGet("actor").(services.UserRead)
		clientIP := middleware.GetClientIP(c)
		userAgent := middleware.GetUserAgent(c)
		actorRoleStr := string(actor.Role)
		_ = auditSvc.Log(c.Request.Context(), services.AuditLogEntry{
			Action:    utils.AuditActionUserList,
			ActorID:   &actor.ID,
			ActorTenantID: &actor.TenantID,
			ActorRole: &actorRoleStr,
			Status:    models.AuditStatusSuccess,
			IPAddress: utils.StringPtr(clientIP),
			UserAgent: utils.StringPtr(userAgent),
			Metadata: map[string]interface{}{
				"resource": "users",
				"tenant_id": req.TenantID,
				"count": len(users),
			},
		})
	}

	c.JSON(http.StatusOK, response)
}

// AdminGetUser gets a specific user by ID
// Requires superadmin role and SYS_USER_LIST permission
func (h *UserHandler) AdminGetUser(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		httpx.HandleError(c, errors.New("user_id is required"))
		return
	}

	user, err := h.authService.AdminGetUser(c.Request.Context(), userID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	response := AdminUserDetail{
		ID:                 user.ID,
		TenantID:           user.TenantID,
		Name:               user.Name,
		Email:              user.Email,
		Role:               string(user.Role),
		ForcePasswordReset: user.ForcePasswordReset,
		CreatedAt:          user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          user.UpdatedAt.Format(time.RFC3339),
		LastLogin:          nil,
	}

	// Audit log
	if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
		actor := c.MustGet("actor").(services.UserRead)
		clientIP := middleware.GetClientIP(c)
		userAgent := middleware.GetUserAgent(c)
		actorRoleStr := string(actor.Role)
		resourceType := "user"
		_ = auditSvc.Log(c.Request.Context(), services.AuditLogEntry{
			Action:             utils.AuditActionUserRead,
			ActorID:            &actor.ID,
			ActorTenantID:      &actor.TenantID,
			ActorRole:          &actorRoleStr,
			TargetResourceID:   &userID,
			TargetResourceType: &resourceType,
			Status:             models.AuditStatusSuccess,
			IPAddress:          utils.StringPtr(clientIP),
			UserAgent:          utils.StringPtr(userAgent),
			Metadata: map[string]interface{}{
				"resource": "user",
				"target_email": user.Email,
			},
		})
	}

	c.JSON(http.StatusOK, response)
}

// AdminResetUserPassword resets a user's password and optionally forces password change
// Requires superadmin role and PLATFORM_USERS_MANAGE permission
func (h *UserHandler) AdminResetUserPassword(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		httpx.HandleError(c, errors.New("user_id is required"))
		return
	}

	var req AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	// Reset password via service
	tempPassword, err := h.authService.AdminResetPassword(c.Request.Context(), userID, req.NewPassword, req.ForceChange)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	// Audit log
	if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
		actor := c.MustGet("actor").(services.UserRead)
		clientIP := middleware.GetClientIP(c)
		userAgent := middleware.GetUserAgent(c)
		actorRoleStr := string(actor.Role)
		resourceType := "user_password"
		_ = auditSvc.Log(c.Request.Context(), services.AuditLogEntry{
			Action:             utils.AuditActionUserPasswordReset,
			ActorID:            &actor.ID,
			ActorTenantID:      &actor.TenantID,
			ActorRole:          &actorRoleStr,
			TargetResourceID:   &userID,
			TargetResourceType: &resourceType,
			Status:             models.AuditStatusSuccess,
			IPAddress:          utils.StringPtr(clientIP),
			UserAgent:          utils.StringPtr(userAgent),
			Metadata: map[string]interface{}{
				"resource": "user_password",
				"force_change": req.ForceChange,
				"temp_password_generated": req.NewPassword == nil,
			},
		})
	}

	response := AdminResetPasswordResponse{
		UserID:             userID,
		TemporaryPassword:  tempPassword,
		ForcePasswordReset: req.ForceChange,
		Message:            "Password has been reset successfully",
	}

	c.JSON(http.StatusOK, response)
}
func (h *UserHandler) resolveTenantBranding(c *gin.Context, tenantID string) *TenantBranding {
	// Fetch tenant settings
	cfg, err := h.tenantSettingService.GetConfiguration(c.Request.Context(), tenantID)
	if err != nil {
		utils.Debug("[resolveTenantBranding] Failed to fetch tenant settings for %s: %v", tenantID, err)
		return nil
	}

	// Extract branding from config
	brandingData, ok := cfg["branding"]
	if !ok {
		utils.Debug("[resolveTenantBranding] No branding found in tenant settings for %s", tenantID)
		return nil
	}

	// Convert to map[string]interface{}
	brandingMap, ok := brandingData.(map[string]interface{})
	if !ok {
		utils.Debug("[resolveTenantBranding] Invalid branding format in tenant settings for %s", tenantID)
		return nil
	}

	branding := &TenantBranding{
		Name:         getStringValue(brandingMap, "name", ""),
		LogoURL:      getStringValue(brandingMap, "logo_url", ""),
		PrimaryColor: getStringValue(brandingMap, "primary_color", "#1F64F0"),
		AccentColor:  getStringValue(brandingMap, "accent_color", "#0D2F81"),
		NavbarTitle:  getStringValue(brandingMap, "navbar_title", ""),
	}

	return branding
}

func resolvePlatformBranding(cfg *config.Config) PlatformBranding {
	// Convert config.PlatformNavItem to handlers.PlatformNavItem
	sidebarLinks := make([]PlatformNavItem, len(cfg.Platform.ParsedSidebarLinks))
	for i, link := range cfg.Platform.ParsedSidebarLinks {
		sidebarLinks[i] = PlatformNavItem{
			Key:   link.Key,
			Label: link.Label,
			Path:  link.Path,
		}
	}

	branding := PlatformBranding{
		Name:         valueOrDefault(cfg.Platform.BrandName, "Recrutr Platform"),
		LogoURL:      valueOrDefault(cfg.Platform.BrandLogoURL, "https://static.recrutr.in/assets/logo.svg"),
		PrimaryColor: valueOrDefault(cfg.Platform.BrandPrimaryColor, "#1F64F0"),
		AccentColor:  valueOrDefault(cfg.Platform.BrandAccentColor, "#0D2F81"),
		NavbarTitle:  valueOrDefault(cfg.Platform.BrandNavbarTitle, "Recrutr Admin"),
		SidebarTitle: valueOrDefault(cfg.Platform.BrandSidebarTitle, "Control Plane"),
		SidebarLinks: sidebarLinks,
	}

	return branding
}

func getStringValue(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func valueOrDefault(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}

func dropClientCache(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
}

// setCookies sets both the access token and CSRF token cookies
func setCookies(c *gin.Context, accessToken string) {
	cfg := config.Get()
	if cfg == nil {
		// Fallback to defaults if config not available
		setAccessTokenCookie(c, accessToken, "", false, "Lax", int((24 * time.Hour).Seconds()))
		setCSRFTokenCookie(c, "", false, "Lax", int((24 * time.Hour).Seconds()))
		return
	}

	cookieCfg := cfg.Cookie
	maxAge := int(cookieCfg.MaxAge.Seconds())

	// Set access token cookie (httpOnly=true, not readable by JS)
	setAccessTokenCookie(c, accessToken, cookieCfg.Domain, cookieCfg.Secure, cookieCfg.SameSite, maxAge)

	// Set CSRF token cookie (httpOnly=false, readable by JS for header)
	setCSRFTokenCookie(c, cookieCfg.Domain, cookieCfg.Secure, cookieCfg.SameSite, maxAge)
}

func setAccessTokenCookie(c *gin.Context, token, domain string, secure bool, sameSite string, maxAge int) {
	c.SetSameSite(parseSameSite(sameSite))
	c.SetCookie(
		"access_token",
		token,
		maxAge,
		"/",
		domain,
		secure,
		true, // httpOnly
	)
}

func setCSRFTokenCookie(c *gin.Context, domain string, secure bool, sameSite string, maxAge int) {
	// Generate CSRF token
	csrfToken, err := utils.GenerateCSRFToken()
	if err != nil {
		utils.Warn("[Cookie] Failed to generate CSRF token: %v", err)
		csrfToken = "fallback-csrf-token" // Fallback (should not happen)
	}

	c.SetSameSite(parseSameSite(sameSite))
	c.SetCookie(
		"csrf_token",
		csrfToken,
		maxAge,
		"/",
		domain,
		secure,
		false, // NOT httpOnly - JS needs to read this
	)
}

func parseSameSite(sameSite string) http.SameSite {
	switch strings.ToLower(sameSite) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	case "lax":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteLaxMode
	}
}

// clearCookies clears both access token and CSRF token cookies
func clearCookies(c *gin.Context) {
	cfg := config.Get()
	domain := ""
	if cfg != nil {
		domain = cfg.Cookie.Domain
	}

	// Clear access_token cookie
	c.SetCookie(
		"access_token",
		"",
		-1, // maxAge=-1 to delete cookie
		"/",
		domain,
		false,
		true,
	)

	// Clear csrf_token cookie
	c.SetCookie(
		"csrf_token",
		"",
		-1, // maxAge=-1 to delete cookie
		"/",
		domain,
		false,
		false,
	)
}
