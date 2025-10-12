package models

// Permission represents a granular authorization capability
type Permission string

// Platform Permissions (Control Plane - SUPERADMIN scope)
const (
	// Tenant Management
	PermTenantList        Permission = "tenant:list"
	PermTenantCreate      Permission = "tenant:create"
	PermTenantRead        Permission = "tenant:read"
	PermTenantUpdate      Permission = "tenant:update"
	PermTenantImpersonate Permission = "tenant:impersonate"
	PermTenantStatus      Permission = "tenant:status"

	// System Management
	PermSysUserList   Permission = "sys:user:list"
	PermSysHealthRead Permission = "sys:health:read"

	// Configuration
	PermSettingsGlobal   Permission = "settings:global"
	PermSettingsSecurity Permission = "settings:security"
	PermSettingsDB       Permission = "settings:db"
)

// Tenant Permissions (Company scope - Tenant-level operations)
const (
	// Analytics
	PermAnalyticsRead Permission = "analytics:read"

	// Jobs (wildcard and granular)
	PermJobAll    Permission = "job:*"
	PermJobCreate Permission = "job:create"
	PermJobRead   Permission = "job:read"
	PermJobUpdate Permission = "job:update"
	PermJobDelete Permission = "job:delete"

	// Applications (wildcard and granular)
	PermApplicationAll    Permission = "application:*"
	PermApplicationCreate Permission = "application:create"
	PermApplicationRead   Permission = "application:read"
	PermApplicationUpdate Permission = "application:update"
	PermApplicationDelete Permission = "application:delete"

	// Pipeline (wildcard and granular)
	PermPipelineAll    Permission = "pipeline:*"
	PermPipelineCreate Permission = "pipeline:create"
	PermPipelineRead   Permission = "pipeline:read"
	PermPipelineUpdate Permission = "pipeline:update"
	PermPipelineDelete Permission = "pipeline:delete"

	// Team Members (wildcard and granular)
	PermMemberAll    Permission = "member:*"
	PermMemberCreate Permission = "member:create"
	PermMemberRead   Permission = "member:read"
	PermMemberUpdate Permission = "member:update"
	PermMemberDelete Permission = "member:delete"

	// Interviews (wildcard and granular)
	PermInterviewAll    Permission = "interview:*"
	PermInterviewCreate Permission = "interview:create"
	PermInterviewRead   Permission = "interview:read"
	PermInterviewUpdate Permission = "interview:update"
	PermInterviewDelete Permission = "interview:delete"

	// Settings (wildcard and granular)
	PermSettingsAll    Permission = "settings:*"
	PermSettingsRead   Permission = "settings:read"
	PermSettingsUpdate Permission = "settings:update"

	// Billing (wildcard and granular)
	PermBillingAll    Permission = "billing:*"
	PermBillingRead   Permission = "billing:read"
	PermBillingUpdate Permission = "billing:update"

	// Integrations (wildcard and granular)
	PermIntegrationsAll    Permission = "integrations:*"
	PermIntegrationsCreate Permission = "integrations:create"
	PermIntegrationsRead   Permission = "integrations:read"
	PermIntegrationsUpdate Permission = "integrations:update"
	PermIntegrationsDelete Permission = "integrations:delete"
)

// HasPermission checks if a user's permissions include the required permission.
// Supports wildcard matching: if user has "job:*", they have "job:create", "job:read", etc.
func HasPermission(userPermissions []string, required string) bool {
	// Check for exact match
	for _, perm := range userPermissions {
		if perm == required {
			return true
		}
	}

	// Check for wildcard match (e.g., user has "job:*", checking for "job:create")
	if len(required) > 0 {
		for _, perm := range userPermissions {
			// Extract namespace from both permission and required
			// Format: "namespace:action" -> check if user has "namespace:*"
			idx := 0
			for i, c := range required {
				if c == ':' {
					idx = i
					break
				}
			}

			if idx > 0 {
				namespace := required[:idx]
				wildcardPerm := namespace + ":*"
				if perm == wildcardPerm {
					return true
				}
			}
		}
	}

	return false
}

// HasAnyPermission checks if user has at least one of the required permissions
func HasAnyPermission(userPermissions []string, required ...string) bool {
	for _, req := range required {
		if HasPermission(userPermissions, req) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if user has all of the required permissions
func HasAllPermissions(userPermissions []string, required ...string) bool {
	for _, req := range required {
		if !HasPermission(userPermissions, req) {
			return false
		}
	}
	return true
}
