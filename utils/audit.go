package utils

import (
	"net/http"
	"strings"
)

// Common audit action constants
const (
	// Authentication actions
	AuditActionLoginSuccess      = "auth.login"
	AuditActionLoginFailed       = "auth.login_failed"
	AuditActionLogout            = "auth.logout"
	AuditActionPasswordChange    = "auth.password_change"
	AuditActionTokenRefresh      = "auth.token_refresh"

	// Tenant actions
	AuditActionTenantCreate      = "tenant.create"
	AuditActionTenantUpdate      = "tenant.update"
	AuditActionTenantDelete      = "tenant.delete"
	AuditActionTenantAccess      = "tenant.access"
	AuditActionTenantAccessDenied = "tenant.access_denied"

	// User actions
	AuditActionUserCreate        = "user.create"
	AuditActionUserUpdate        = "user.update"
	AuditActionUserDelete        = "user.delete"
	AuditActionUserList          = "user.list"

	// Subscription actions
	AuditActionSubscriptionCreate  = "subscription.create"
	AuditActionSubscriptionUpdate  = "subscription.update"
	AuditActionSubscriptionCancel  = "subscription.cancel"
	AuditActionSubscriptionSuspend = "subscription.suspend"

	// Permission actions
	AuditActionPermissionDenied  = "permission.denied"

	// Impersonation actions
	AuditActionImpersonationStart = "impersonation.start"
	AuditActionImpersonationEnd   = "impersonation.end"
)

// ExtractClientIP extracts the client IP address from the request
// Considers X-Forwarded-For, X-Real-IP headers and RemoteAddr
func ExtractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (may contain comma-separated list)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	// RemoteAddr format is "IP:port", so we need to strip the port
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}

	return r.RemoteAddr
}

// ExtractUserAgent extracts the User-Agent header from the request
func ExtractUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// StringPtr converts a string to a pointer
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
