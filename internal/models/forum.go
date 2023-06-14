package models

// easyjson -all ./internal/models/forum.go

type Forum struct {
	ID      int    `json:"-"`
	Title   string `json:"title"`
	User    string `json:"user"`
	Slug    string `json:"slug"`
	Posts   int    `json:"posts,omitempty"`
	Threads int    `json:"threads,omitempty"`
}
