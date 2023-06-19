package models

// easyjson -all ./internal/models/status.go

type Status struct {
	Forums  int64 `json:"forum"`
	Posts   int64 `json:"post"`
	Threads int64 `json:"thread"`
	Users   int64 `json:"user"`
}
