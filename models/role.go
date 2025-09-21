package models

type Role string

const (
	RoleAdmin       Role = "ADMIN"
	RoleHR          Role = "HR"
	RoleInterviewer Role = "INTERVIEWER"
	RoleCandidate   Role = "CANDIDATE"
)
