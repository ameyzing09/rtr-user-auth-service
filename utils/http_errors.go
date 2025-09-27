package utils

import (
	"errors"
	"net/http"

	"rtr-user-auth-service/domain"
	errcodes "rtr-user-auth-service/errors"
)

func ResolveHTTPError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrSuperadminRequired):
		return http.StatusForbidden, errcodes.ErrCodeSuperadminRequired
	case errors.Is(err, domain.ErrTenantSlugTaken):
		return http.StatusConflict, errcodes.ErrCodeTenantSlugTaken
	case errors.Is(err, domain.ErrDomainInUse), errors.Is(err, domain.ErrEmailInUse):
		return http.StatusConflict, errcodes.ErrCodeDomainInUse
	case errors.Is(err, domain.ErrIdempotencyKeyReuseDifferentReq):
		return http.StatusConflict, errcodes.ErrCodeIdempotencyKeyReuseDiff
	case errors.Is(err, domain.ErrTenantNotFound):
		return http.StatusNotFound, errcodes.ErrCodeTenantNotFound
	default:
		return http.StatusInternalServerError, errcodes.ErrCodeInternal
	}
}
