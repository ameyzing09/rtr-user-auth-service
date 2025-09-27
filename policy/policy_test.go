package policy

import (
	"rtr-user-auth-service/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCan(t *testing.T) {
	tests := []struct {
		name     string
		role     models.Role
		action   Action
		expected bool
	}{
		// SUPERADMIN can do everything
		{
			name:     "SUPERADMIN can list users",
			role:     models.RoleSuperAdmin,
			action:   ActionUserList,
			expected: true,
		},
		{
			name:     "SUPERADMIN can create tenants",
			role:     models.RoleSuperAdmin,
			action:   ActionTenantCreate,
			expected: true,
		},
		{
			name:     "SUPERADMIN can delete users",
			role:     models.RoleSuperAdmin,
			action:   ActionUserDelete,
			expected: true,
		},

		// ADMIN permissions
		{
			name:     "ADMIN can list users",
			role:     models.RoleAdmin,
			action:   ActionUserList,
			expected: true,
		},
		{
			name:     "ADMIN can create users",
			role:     models.RoleAdmin,
			action:   ActionUserCreate,
			expected: true,
		},
		{
			name:     "ADMIN can update users",
			role:     models.RoleAdmin,
			action:   ActionUserUpdate,
			expected: true,
		},
		{
			name:     "ADMIN cannot delete users",
			role:     models.RoleAdmin,
			action:   ActionUserDelete,
			expected: false,
		},
		{
			name:     "ADMIN can update tenant settings",
			role:     models.RoleAdmin,
			action:   ActionTenantSettingsUpdate,
			expected: true,
		},
		{
			name:     "ADMIN cannot create tenants",
			role:     models.RoleAdmin,
			action:   ActionTenantCreate,
			expected: false,
		},

		// HR permissions
		{
			name:     "HR can list users",
			role:     models.RoleHR,
			action:   ActionUserList,
			expected: true,
		},
		{
			name:     "HR can create users",
			role:     models.RoleHR,
			action:   ActionUserCreate,
			expected: true,
		},
		{
			name:     "HR cannot update users",
			role:     models.RoleHR,
			action:   ActionUserUpdate,
			expected: false,
		},
		{
			name:     "HR cannot delete users",
			role:     models.RoleHR,
			action:   ActionUserDelete,
			expected: false,
		},
		{
			name:     "HR can view tenant settings",
			role:     models.RoleHR,
			action:   ActionTenantSettingsView,
			expected: true,
		},
		{
			name:     "HR cannot update tenant settings",
			role:     models.RoleHR,
			action:   ActionTenantSettingsUpdate,
			expected: false,
		},

		// INTERVIEWER permissions
		{
			name:     "INTERVIEWER can view profile",
			role:     models.RoleInterviewer,
			action:   ActionProfileView,
			expected: true,
		},
		{
			name:     "INTERVIEWER can update profile",
			role:     models.RoleInterviewer,
			action:   ActionProfileUpdate,
			expected: true,
		},
		{
			name:     "INTERVIEWER cannot list users",
			role:     models.RoleInterviewer,
			action:   ActionUserList,
			expected: false,
		},
		{
			name:     "INTERVIEWER cannot create users",
			role:     models.RoleInterviewer,
			action:   ActionUserCreate,
			expected: false,
		},

		// CANDIDATE permissions
		{
			name:     "CANDIDATE can view profile",
			role:     models.RoleCandidate,
			action:   ActionProfileView,
			expected: true,
		},
		{
			name:     "CANDIDATE can update profile",
			role:     models.RoleCandidate,
			action:   ActionProfileUpdate,
			expected: true,
		},
		{
			name:     "CANDIDATE cannot list users",
			role:     models.RoleCandidate,
			action:   ActionUserList,
			expected: false,
		},
		{
			name:     "CANDIDATE cannot create users",
			role:     models.RoleCandidate,
			action:   ActionUserCreate,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Can(tt.role, tt.action)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCanAny(t *testing.T) {
	tests := []struct {
		name     string
		role     models.Role
		actions  []Action
		expected bool
	}{
		{
			name:     "ADMIN can do any of user management actions",
			role:     models.RoleAdmin,
			actions:  []Action{ActionUserList, ActionUserCreate, ActionUserUpdate},
			expected: true,
		},
		{
			name:     "HR can do some user management actions",
			role:     models.RoleHR,
			actions:  []Action{ActionUserList, ActionUserCreate, ActionUserUpdate},
			expected: true,
		},
		{
			name:     "INTERVIEWER cannot do any user management actions",
			role:     models.RoleInterviewer,
			actions:  []Action{ActionUserList, ActionUserCreate, ActionUserUpdate},
			expected: false,
		},
		{
			name:     "CANDIDATE can do profile actions",
			role:     models.RoleCandidate,
			actions:  []Action{ActionProfileView, ActionProfileUpdate},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanAny(tt.role, tt.actions...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCanAll(t *testing.T) {
	tests := []struct {
		name     string
		role     models.Role
		actions  []Action
		expected bool
	}{
		{
			name:     "ADMIN can do all user management actions",
			role:     models.RoleAdmin,
			actions:  []Action{ActionUserList, ActionUserCreate, ActionUserUpdate},
			expected: true,
		},
		{
			name:     "ADMIN cannot do all actions including delete",
			role:     models.RoleAdmin,
			actions:  []Action{ActionUserList, ActionUserCreate, ActionUserDelete},
			expected: false,
		},
		{
			name:     "HR cannot do all user management actions",
			role:     models.RoleHR,
			actions:  []Action{ActionUserList, ActionUserCreate, ActionUserUpdate},
			expected: false,
		},
		{
			name:     "CANDIDATE can do all profile actions",
			role:     models.RoleCandidate,
			actions:  []Action{ActionProfileView, ActionProfileUpdate},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanAll(tt.role, tt.actions...)
			assert.Equal(t, tt.expected, result)
		})
	}
}
