package errcodes

const (
	ErrCodeSuperadminRequired      = "SUPERADMIN_REQUIRED"
	ErrCodeTenantSlugTaken         = "TENANT_SLUG_TAKEN"
	ErrCodeDomainInUse             = "DOMAIN_IN_USE"
	ErrCodeValidation              = "VALIDATION_ERROR"
	ErrCodeIdempotencyKeyReuseDiff = "IDEMPOTENCY_KEY_REUSE_DIFFERENT_REQUEST"
	ErrCodeTenantNotFound          = "TENANT_NOT_FOUND"
	ErrCodeInternal                = "INTERNAL_ERROR"
)
