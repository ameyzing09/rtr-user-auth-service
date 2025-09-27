package services

import "errors"

var (
	ErrInvalidCredentials            = errors.New("invalid credentials")
	ErrForbidden                     = errors.New("forbidden")
	ErrUserNotFound                  = errors.New("user not found")
	ErrEmailInUse                    = errors.New("email already in use")
	ErrInvalidInput                  = errors.New("invalid input")
	ErrDomainInUse                   = errors.New("domain already in use")
	ErrTenantSlugTaken               = errors.New("tenant slug already taken")
	ErrIdempotencyKeyConflict        = errors.New("idempotency key reuse with different request")
	ErrIdempotencyProcessingConflict = errors.New("idempotency record busy")
	ErrSuperadminRequired            = errors.New("superadmin required")
)
