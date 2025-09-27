package policy

import "rtr-user-auth-service/models"

// Action represents a specific action that can be performed
type Action string

const (
	// User management actions
	ActionUserList   Action = "user:list"
	ActionUserCreate Action = "user:create"
	ActionUserUpdate Action = "user:update"
	ActionUserDelete Action = "user:delete"

	// Tenant management actions
	ActionTenantCreate Action = "tenant:create"
	ActionTenantUpdate Action = "tenant:update"
	ActionTenantDelete Action = "tenant:delete"
	ActionTenantView   Action = "tenant:view"

	// Tenant settings actions
	ActionTenantSettingsView   Action = "tenant_settings:view"
	ActionTenantSettingsUpdate Action = "tenant_settings:update"

	// Profile actions
	ActionProfileView   Action = "profile:view"
	ActionProfileUpdate Action = "profile:update"
)

// Can checks if a role can perform a specific action
func Can(role models.Role, action Action) bool {
	switch role {
	case models.RoleSuperAdmin:
		// SUPERADMIN can do everything
		return true

	case models.RoleAdmin:
		// ADMIN can manage users and tenant settings, but not delete users
		switch action {
		case ActionUserList, ActionUserCreate, ActionUserUpdate,
			ActionTenantSettingsView, ActionTenantSettingsUpdate,
			ActionProfileView, ActionProfileUpdate:
			return true
		case ActionUserDelete, ActionTenantCreate, ActionTenantUpdate, ActionTenantDelete:
			return false
		default:
			return false
		}

	case models.RoleHR:
		// HR can list and create users, view tenant settings
		switch action {
		case ActionUserList, ActionUserCreate,
			ActionTenantSettingsView,
			ActionProfileView, ActionProfileUpdate:
			return true
		case ActionUserUpdate, ActionUserDelete,
			ActionTenantSettingsUpdate,
			ActionTenantCreate, ActionTenantUpdate, ActionTenantDelete:
			return false
		default:
			return false
		}

	case models.RoleInterviewer:
		// INTERVIEWER can only view their profile
		switch action {
		case ActionProfileView, ActionProfileUpdate:
			return true
		default:
			return false
		}

	case models.RoleCandidate:
		// CANDIDATE can only view their profile
		switch action {
		case ActionProfileView, ActionProfileUpdate:
			return true
		default:
			return false
		}

	default:
		return false
	}
}

// CanAny checks if a role can perform any of the specified actions
func CanAny(role models.Role, actions ...Action) bool {
	for _, action := range actions {
		if Can(role, action) {
			return true
		}
	}
	return false
}

// CanAll checks if a role can perform all of the specified actions
func CanAll(role models.Role, actions ...Action) bool {
	for _, action := range actions {
		if !Can(role, action) {
			return false
		}
	}
	return true
}
