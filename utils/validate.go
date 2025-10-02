package utils

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
)

// domainPattern allows single-label domains (for dev environments) and multi-label domains.
var domainPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]*[a-z0-9])?)*$`)

func NormalizeEmail(email string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(email))
	if value == "" {
		return "", errors.New("email cannot be empty")
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return "", err
	}
	return value, nil
}

func NormalizeDomain(domain string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(domain))
	if value == "" {
		return "", errors.New("domain cannot be empty")
	}

	value = strings.TrimPrefix(value, "https://")
	value = strings.TrimPrefix(value, "http://")
	if idx := strings.Index(value, "/"); idx > -1 {
		value = value[:idx]
	}
	if idx := strings.Index(value, ":"); idx > -1 {
		value = value[:idx]
	}
	value = strings.Trim(value, ".")

	if len(value) < 3 || len(value) > 253 {
		return "", errors.New("domain length invalid")
	}

	if !domainPattern.MatchString(value) {
		return "", errors.New("domain format invalid")
	}

	return value, nil
}
