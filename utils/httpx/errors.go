package httpx

import (
	"net/http"
	"rtr-user-auth-service/domain"

	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, err error) {
	switch err {
	case domain.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
	case domain.ErrUnauthorized, domain.ErrInvalidCredentials:
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
	case domain.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case domain.ErrEmailInUse:
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	case domain.ErrTenantNotFound:
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
}
