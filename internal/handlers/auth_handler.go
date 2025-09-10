package handlers

import (
	"net/http"

	"github.com/ameyzing09/rtr-user-auth-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService services.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	TenantID uuid.UUID `json:"tenant_id" binding:"required"`
	Email    string    `json:"email" binding:"required,email"`
	Password string    `json:"password" binding:"required,min=8"`
}

// RefreshTokenRequest represents the refresh token request body
type RefreshTokenRequest struct {
	TenantID     uuid.UUID `json:"tenant_id" binding:"required"`
	RefreshToken string    `json:"refresh_token" binding:"required"`
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} services.LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	response, err := h.authService.Login(c.Request.Context(), req.TenantID, req.Email, req.Password)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_credentials",
				"message": "Invalid email or password",
			})
		case services.ErrUserInactive:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "user_inactive",
				"message": "User account is inactive",
			})
		case services.ErrTenantInactive:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "tenant_inactive",
				"message": "Tenant account is inactive",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "An internal error occurred",
			})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} services.LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	response, err := h.authService.RefreshToken(c.Request.Context(), req.TenantID, req.RefreshToken)
	if err != nil {
		switch err {
		case services.ErrInvalidRefreshToken, services.ErrTokenExpired:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_refresh_token",
				"message": "Invalid or expired refresh token",
			})
		case services.ErrUserNotFound:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "user_not_found",
				"message": "User not found",
			})
		case services.ErrUserInactive:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "user_inactive",
				"message": "User account is inactive",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "An internal error occurred",
			})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout godoc
// @Summary User logout
// @Description Logout user and revoke refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	tenantID, exists := c.Get("tenant")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Tenant not found in context",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Invalid user ID type",
		})
		return
	}

	tenantUUID, ok := tenantID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Invalid tenant ID type",
		})
		return
	}

	err := h.authService.Logout(c.Request.Context(), tenantUUID, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to logout user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}