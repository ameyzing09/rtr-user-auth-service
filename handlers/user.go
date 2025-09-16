package handlers

import (
	"net/http"
	"rtr-user-auth-service/services"
	"rtr-user-auth-service/utils/httpx"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	authService services.AuthService
}

func NewUserHandler(authService services.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

func (h *UserHandler) Register(c *gin.Context) {
	var registerReq RegisterRequest
	if err := c.ShouldBindBodyWithJSON((&registerReq)); err != nil {
		c.JSON((http.StatusBadRequest), gin.H{"error": err.Error()})
		return
	}
	actor := c.MustGet("actor").(services.UserRead)

	output, err := h.authService.Register(c, actor, services.RegisterInput{
		TenantID: actor.TenantID,
		Name:     registerReq.Name,
		Email:    registerReq.Email,
		Password: registerReq.Password,
		Role:     registerReq.Role,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, output)
}

func (h *UserHandler) Login(c *gin.Context) {
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON((http.StatusBadRequest), gin.H{"error": err.Error()})
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

	c.JSON(http.StatusOK, gin.H{
		"Token":     token.Token,
		"ExpiresAt": token.ExpiresAt,
		"User":      user,
	})
}

func (h *UserHandler) GetMe(c *gin.Context) {
	actor := c.MustGet("actor").(services.UserRead)
	users, err := h.authService.GetMe(c, actor.ID, actor.TenantID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, users)
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
