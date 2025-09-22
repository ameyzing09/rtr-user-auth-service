package services

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrForbidden           = errors.New("forbidden")
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailInUse          = errors.New("email already in use")
	ErrInvalidInput        = errors.New("invalid input")
	ErrDomainExists        = errors.New("domain already exists")
	ErrTenantAlreadyExists = errors.New("tenant already exists")
)
