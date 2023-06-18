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
	Path     pgtype.Int4Array `json:"path,omitempty"`
}

// easyjson:skip
type SortParams struct {
	Limit string
	Since string
	Desc  string
	Sort  string
}

type PostFull struct {
	Thread *Thread `json:"thread,omitempty"`
	Forum  *Forum  `json:"forum,omitempty"`
	Author *User   `json:"author,omitempty"`
	Post   Post    `json:"post,omitempty"`
}

type PostUpdate struct {
	ID      int    `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}
