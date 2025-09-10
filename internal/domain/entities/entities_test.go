package entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want bool
	}{
		{
			name: "valid admin role",
			role: RoleAdmin,
			want: true,
		},
		{
			name: "valid hr role",
			role: RoleHR,
			want: true,
		},
		{
			name: "valid interviewer role",
			role: RoleInterviewer,
			want: true,
		},
		{
			name: "valid candidate role",
			role: RoleCandidate,
			want: true,
		},
		{
			name: "invalid role",
			role: Role("INVALID"),
			want: false,
		},
		{
			name: "empty role",
			role: Role(""),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.role.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUser_GetFullName(t *testing.T) {
	user := &User{
		FirstName: "John",
		LastName:  "Doe",
	}

	fullName := user.GetFullName()
	assert.Equal(t, "John Doe", fullName)
}

func TestUser_BeforeCreate(t *testing.T) {
	user := &User{}
	
	// ID should be nil initially
	assert.Equal(t, uuid.Nil, user.ID)
	
	// BeforeCreate should generate UUID
	err := user.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
}

func TestTenant_BeforeCreate(t *testing.T) {
	tenant := &Tenant{}
	
	// ID should be nil initially
	assert.Equal(t, uuid.Nil, tenant.ID)
	
	// BeforeCreate should generate UUID
	err := tenant.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, tenant.ID)
}

func TestRefreshToken_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		token    *RefreshToken
		expected bool
	}{
		{
			name: "expired token",
			token: &RefreshToken{
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			},
			expected: true,
		},
		{
			name: "valid token",
			token: &RefreshToken{
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRefreshToken_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		token    *RefreshToken
		expected bool
	}{
		{
			name: "valid token",
			token: &RefreshToken{
				ExpiresAt: time.Now().Add(1 * time.Hour),
				IsRevoked: false,
			},
			expected: true,
		},
		{
			name: "expired token",
			token: &RefreshToken{
				ExpiresAt: time.Now().Add(-1 * time.Hour),
				IsRevoked: false,
			},
			expected: false,
		},
		{
			name: "revoked token",
			token: &RefreshToken{
				ExpiresAt: time.Now().Add(1 * time.Hour),
				IsRevoked: true,
			},
			expected: false,
		},
		{
			name: "expired and revoked token",
			token: &RefreshToken{
				ExpiresAt: time.Now().Add(-1 * time.Hour),
				IsRevoked: true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}