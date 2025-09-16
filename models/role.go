package models

type Role string

const (
	RoleAdmin       Role = "admin"
	RoleHR          Role = "hr"
	RoleInterviewer Role = "interviewer"
	RoleCandidate   Role = "candidate"
)
