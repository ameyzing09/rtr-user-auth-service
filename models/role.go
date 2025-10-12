package models

type Role string

const (
	RoleSuperAdmin  Role = "SUPERADMIN"
	RoleAdmin       Role = "ADMIN" // Note: Frontend should use TENANT_ADMIN for clarity, but DB still uses ADMIN
	RoleHR          Role = "HR"
	RoleInterviewer Role = "INTERVIEWER"
	RoleViewer      Role = "VIEWER"
	RoleCandidate   Role = "CANDIDATE"
)
