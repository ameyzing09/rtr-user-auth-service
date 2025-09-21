package httpx

import (
	"net/http"
	"rtr-user-auth-service/domain"

	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, err error) {
	switch err {
	case domain.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case domain.ErrUnauthorized, domain.ErrInvalidCredentials:
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case domain.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case domain.ErrEmailInUse:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case domain.ErrTenantNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
