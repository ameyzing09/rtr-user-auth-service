package models

// RolePermissions maps each role to its granted permissions
// This is the single source of truth for role-based permission assignment
var RolePermissions = map[Role][]string{
	// SUPERADMIN: All platform/control-plane permissions
	// Note: SUPERADMIN does NOT automatically get tenant-feature permissions
	// unless they impersonate or are explicitly granted tenant access
	RoleSuperAdmin: {
		// Tenant Management
		string(PermTenantList),
		string(PermTenantCreate),
		string(PermTenantRead),
		string(PermTenantUpdate),
		string(PermTenantImpersonate),
		string(PermTenantStatus),

		// System Management
		string(PermSysUserList),
		string(PermSysHealthRead),

		// Analytics (platform-level)
		string(PermAnalyticsRead),

		// Configuration
		string(PermSettingsGlobal),
		string(PermSettingsSecurity),
		string(PermSettingsDB),
	},

	// TENANT_ADMIN (formerly ADMIN): All tenant-scoped permissions
	// Has full control over tenant resources but no platform/system access
	RoleAdmin: {
		// Analytics
		string(PermAnalyticsRead),

		// All tenant feature permissions (using wildcards)
		string(PermJobAll),
		string(PermApplicationAll),
		string(PermPipelineAll),
		string(PermMemberAll),
		string(PermInterviewAll),
		string(PermSettingsAll),
		string(PermBillingAll),
		string(PermIntegrationsAll),
		string(PermFeedbackAll),

		// Explicit list permissions for granular control
		string(PermJobList),
		string(PermApplicationList),
		string(PermPipelineList),
		string(PermMemberList),
		string(PermInterviewList),
	},

	// HR: Job posting, applications, pipeline, team member, and interview management
	RoleHR: {
		string(PermJobAll),
		string(PermApplicationAll),
		string(PermPipelineAll),
		string(PermMemberAll),
		string(PermInterviewAll), // HR schedules and manages interviews
		string(PermFeedbackAll),  // HR can manage interview feedback

		// Explicit list permissions
		string(PermJobList),
		string(PermApplicationList),
		string(PermPipelineList),
		string(PermMemberList),
		string(PermInterviewList),
	},

	// INTERVIEWER: Can manage interviews and view applications
	RoleInterviewer: {
		string(PermInterviewAll),
		string(PermFeedbackAll), // Interviewers provide feedback
		string(PermApplicationRead),

		// List permissions for interviews and applications
		string(PermInterviewList),
		string(PermApplicationList),
	},

	// VIEWER: Deprecated role - no permissions
	RoleViewer: {},

	// CANDIDATE: Read-only access to analytics and applications
	// Row-level security ensures candidates only see their own data
	RoleCandidate: {
		string(PermAnalyticsRead),
		string(PermApplicationRead),
		string(PermApplicationList),
	},
}

// GetRolePermissions returns the permissions for a given role
func GetRolePermissions(role Role) []string {
	if perms, ok := RolePermissions[role]; ok {
		// Return a copy to prevent external modification
		result := make([]string, len(perms))
		copy(result, perms)
		return result
	}
	return []string{}
}

// RoleHasPermission checks if a role has a specific permission
func RoleHasPermission(role Role, permission string) bool {
	perms := GetRolePermissions(role)
	return HasPermission(perms, permission)
}

// ValidateRolePermissions ensures all roles have valid permission mappings
// This can be used in tests or during application startup
func ValidateRolePermissions() error {
	requiredRoles := []Role{
		RoleSuperAdmin,
		RoleAdmin,
		RoleHR,
		RoleInterviewer,
		RoleViewer,
		RoleCandidate,
	}

	for _, role := range requiredRoles {
		if _, ok := RolePermissions[role]; !ok {
			return ErrInvalidRole(string(role))
		}
	}

	return nil
}

// ErrInvalidRole is returned when a role doesn't have permission mapping
func ErrInvalidRole(role string) error {
	return &InvalidRoleError{Role: role}
}

type InvalidRoleError struct {
	Role string
}

func (e *InvalidRoleError) Error() string {
	return "invalid role: " + e.Role
}
