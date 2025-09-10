package handlers

import (
	"net/http"
	"strconv"

	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/entities"
	"github.com/ameyzing09/rtr-user-auth-service/internal/services"
	"github.com/ameyzing09/rtr-user-auth-service/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler handles user-related endpoints
type UserHandler struct {
	userService services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Email     string        `json:"email" binding:"required,email"`
	Password  string        `json:"password" binding:"required,min=8"`
	FirstName string        `json:"first_name" binding:"required,min=2,max=50"`
	LastName  string        `json:"last_name" binding:"required,min=2,max=50"`
	Role      entities.Role `json:"role" binding:"required"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Email     *string        `json:"email,omitempty"`
	FirstName *string        `json:"first_name,omitempty"`
	LastName  *string        `json:"last_name,omitempty"`
	Role      *entities.Role `json:"role,omitempty"`
	IsActive  *bool          `json:"is_active,omitempty"`
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user within a tenant
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantId path string true "Tenant ID"
// @Param request body CreateUserRequest true "User details"
// @Success 201 {object} entities.User
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tenants/{tenantId}/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	tenantIDParam := c.Param("tenantId")
	tenantID, err := uuid.Parse(tenantIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid tenant ID format",
		})
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	serviceReq := &services.CreateUserRequest{
		TenantID:  tenantID,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
	}

	user, err := h.userService.Create(c.Request.Context(), serviceReq)
	if err != nil {
		switch err {
		case services.ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{
				"error":   "user_already_exists",
				"message": "User with this email already exists",
			})
		case services.ErrInvalidRole:
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_role",
				"message": "Invalid role specified",
			})
		case services.ErrTenantNotFound:
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "tenant_not_found",
				"message": "Tenant not found",
			})
		case services.ErrTenantInactive:
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "tenant_inactive",
				"message": "Tenant is inactive",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to create user",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get user details by ID within a tenant
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantId path string true "Tenant ID"
// @Param userId path string true "User ID"
// @Success 200 {object} entities.User
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tenants/{tenantId}/users/{userId} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	tenantIDParam := c.Param("tenantId")
	tenantID, err := uuid.Parse(tenantIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid tenant ID format",
		})
		return
	}

	userIDParam := c.Param("userId")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid user ID format",
		})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), tenantID, userID)
	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user_not_found",
				"message": "User not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to get user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user details
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantId path string true "Tenant ID"
// @Param userId path string true "User ID"
// @Param request body UpdateUserRequest true "Updated user details"
// @Success 200 {object} entities.User
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tenants/{tenantId}/users/{userId} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	tenantIDParam := c.Param("tenantId")
	tenantID, err := uuid.Parse(tenantIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid tenant ID format",
		})
		return
	}

	userIDParam := c.Param("userId")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid user ID format",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	serviceReq := &services.UpdateUserRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
		IsActive:  req.IsActive,
	}

	user, err := h.userService.Update(c.Request.Context(), tenantID, userID, serviceReq)
	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user_not_found",
				"message": "User not found",
			})
		case services.ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{
				"error":   "email_already_exists",
				"message": "Email already exists",
			})
		case services.ErrInvalidRole:
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_role",
				"message": "Invalid role specified",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to update user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete user from tenant
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantId path string true "Tenant ID"
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tenants/{tenantId}/users/{userId} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	tenantIDParam := c.Param("tenantId")
	tenantID, err := uuid.Parse(tenantIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid tenant ID format",
		})
		return
	}

	userIDParam := c.Param("userId")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid user ID format",
		})
		return
	}

	err = h.userService.Delete(c.Request.Context(), tenantID, userID)
	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user_not_found",
				"message": "User not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to delete user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// ListUsers godoc
// @Summary List users
// @Description List users within a tenant with pagination
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantId path string true "Tenant ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param role query string false "Filter by role"
// @Success 200 {object} services.ListUsersResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tenants/{tenantId}/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	tenantIDParam := c.Param("tenantId")
	tenantID, err := uuid.Parse(tenantIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid tenant ID format",
		})
		return
	}

	// Parse query parameters
	page := 1
	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	roleParam := c.Query("role")
	
	req := &services.ListUsersRequest{
		Page:  page,
		Limit: limit,
	}

	var response *services.ListUsersResponse

	if roleParam != "" {
		role := entities.Role(roleParam)
		if !role.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_role",
				"message": "Invalid role filter",
			})
			return
		}
		response, err = h.userService.ListByRole(c.Request.Context(), tenantID, role, req)
	} else {
		response, err = h.userService.List(c.Request.Context(), tenantID, req)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to list users",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Get the profile of the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} entities.User
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	userClaims, ok := claims.(*utils.JWTClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Invalid claims type",
		})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userClaims.TenantID, userClaims.UserID)
	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user_not_found",
				"message": "User not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to get user profile",
			})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}