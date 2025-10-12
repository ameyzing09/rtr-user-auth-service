package handlers

import (
	"net/http"
	"strings"
	"time"

	"rtr-user-auth-service/config"
	"rtr-user-auth-service/domain"
	errcodes "rtr-user-auth-service/errors"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/services"
	"rtr-user-auth-service/utils"
	"rtr-user-auth-service/utils/httpx"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	authService services.AuthService
}

func NewUserHandler(authService services.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

func (h *UserHandler) Login(c *gin.Context) {
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	token, user, err := h.authService.Login(c, services.LoginInput{
		Email:    loginReq.Email,
		Password: loginReq.Password,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	c.Header("X-Tenant-ID", user.TenantID)

	// Set httpOnly cookie for JWT token
	setCookies(c, token.Token)

	response := LoginResponse{
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
		User:      user,
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) AdminLogin(c *gin.Context) {
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	token, user, err := h.authService.Login(c, services.LoginInput{
		Email:    loginReq.Email,
		Password: loginReq.Password,
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
	// Clear cookies
	clearCookies(c)
	dropClientCache(c)
	c.Status(http.StatusNoContent)
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
