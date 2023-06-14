package models

import (
	"github.com/jackc/pgtype"
	"time"
)

// easyjson -all ./internal/models/post.go

type Post struct {
	ID       int              `json:"id,omitempty"`
	Parent   int              `json:"parent,omitempty"`
	Author   string           `json:"author"`
	Message  string           `json:"message"`
	IsEdited bool             `json:"isEdited,omitempty"`
	Forum    string           `json:"forum,omitempty"`
	Thread   int              `json:"thread,omitempty"`
	Created  time.Time        `json:"created,omitempty"`
	Path     pgtype.Int8Array `json:"path,omitempty"`
}
