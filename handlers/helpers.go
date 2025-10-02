package handlers

import (
	"net/http"
	"strings"

	errcodes "rtr-user-auth-service/errors"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/services"

	"github.com/gin-gonic/gin"
)

const actorContextKey = "actor"

// GetActorFromContext retrieves the authenticated actor from the Gin context
// Returns the actor and true if found, or responds with error and returns false
func GetActorFromContext(c *gin.Context) (services.UserRead, bool) {
	actorValue, exists := c.Get(actorContextKey)
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    errcodes.ErrCodeInternal,
			"message": "authentication context missing",
		})
		return services.UserRead{}, false
	}

	actor, ok := actorValue.(services.UserRead)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    errcodes.ErrCodeInternal,
			"message": "authentication context invalid",
		})
		return services.UserRead{}, false
	}

	return actor, true
}

// ActorFromContext retrieves the authenticated actor from the Gin context
// Panics if actor is not found (assumes middleware has validated it)
func ActorFromContext(c *gin.Context) services.UserRead {
	return c.MustGet(actorContextKey).(services.UserRead)
}

// StringPointer converts a string pointer to a non-nil pointer value
// Returns nil if input is nil or empty string after trimming
func StringPointer(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// StringValue safely dereferences a string pointer, returning empty string if nil
func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// PlanPointer converts a string pointer to a Plan pointer
// Returns nil if input is nil or empty after trimming
func PlanPointer(plan *string) *models.Plan {
	if plan == nil {
		return nil
	}
	trimmed := strings.ToUpper(strings.TrimSpace(*plan))
	if trimmed == "" {
		return nil
	}
	value := models.Plan(trimmed)
	return &value
}

// PlanStringPointer converts a Plan pointer to a string pointer
func PlanStringPointer(plan *models.Plan) *string {
	if plan == nil {
		return nil
	}
	value := string(*plan)
	return &value
}

// CopyStringPointer creates a copy of a string pointer to avoid aliasing issues
func CopyStringPointer(s *string) *string {
	if s == nil {
		return nil
	}
	value := *s
	return &value
}
