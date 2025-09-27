package domain

import "errors"

var (
	ErrUnauthorized                    = errors.New("unauthorized")
	ErrForbidden                       = errors.New("forbidden")
	ErrInvalidCredentials              = errors.New("invalid credentials")
	ErrEmailInUse                      = errors.New("email already in use for tenant")
	ErrUserNotFound                    = errors.New("user not found")
	ErrTenantNotFound                  = errors.New("tenant not found")
	ErrDomainInUse                     = errors.New("domain already in use")
	ErrTenantSlugTaken                 = errors.New("tenant slug already taken")
	ErrSuperadminRequired              = errors.New("superadmin required")
	ErrIdempotencyKeyReuseDifferentReq = errors.New("idempotency key reuse with different request")
)
