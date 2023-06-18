package models

// easyjson -all ./internal/models/vote.go

type Vote struct {
	Nickname string `json:"nickname"`
	Voice    int    `json:"voice"`
	Thread   int    `json:"-"`
}
