package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

const (
	// CSRFTokenLength is the length of CSRF tokens in bytes (32 bytes = 256 bits)
	CSRFTokenLength = 32
)

// GenerateCSRFToken generates a cryptographically secure random CSRF token
func GenerateCSRFToken() (string, error) {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ValidateCSRFToken validates that the token from header matches the token from cookie
// Returns true if tokens match, false otherwise
func ValidateCSRFToken(headerToken, cookieToken string) bool {
	if headerToken == "" || cookieToken == "" {
		return false
	}
	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(headerToken), []byte(cookieToken)) == 1
}
