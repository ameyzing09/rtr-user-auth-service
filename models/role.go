package models

type Role string

const (
	RoleSuperAdmin  Role = "SUPERADMIN"
	RoleAdmin       Role = "ADMIN"
	RoleHR          Role = "HR"
	RoleInterviewer Role = "INTERVIEWER"
	RoleCandidate   Role = "CANDIDATE"
)
