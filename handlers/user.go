package handlers

import (
	"net/http"
	"os"
	"strings"

	"rtr-user-auth-service/domain"
	errcodes "rtr-user-auth-service/errors"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/services"
	"rtr-user-auth-service/utils/httpx"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	authService services.AuthService
}

var defaultPlatformBranding = PlatformBranding{
	Name:         "Recrutr Platform",
	LogoURL:      "https://static.recrutr.in/assets/logo.svg",
	PrimaryColor: "#1F64F0",
	AccentColor:  "#0D2F81",
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

	branding := resolvePlatformBranding()

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

func (h *UserHandler) Logout(c *gin.Context) {
	dropClientCache(c)
	c.Status(http.StatusNoContent)
}

func resolvePlatformBranding() PlatformBranding {
	name := strings.TrimSpace(os.Getenv("PLATFORM_BRAND_NAME"))
	logoURL := strings.TrimSpace(os.Getenv("PLATFORM_BRAND_LOGO_URL"))
	primary := strings.TrimSpace(os.Getenv("PLATFORM_BRAND_PRIMARY_COLOR"))
	accent := strings.TrimSpace(os.Getenv("PLATFORM_BRAND_ACCENT_COLOR"))

	branding := defaultPlatformBranding

	if name != "" {
		branding.Name = name
	}
	if logoURL != "" {
		branding.LogoURL = logoURL
	}
	if primary != "" {
		branding.PrimaryColor = primary
	}
	if accent != "" {
		branding.AccentColor = accent
	}

	return branding
}

func dropClientCache(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
}
