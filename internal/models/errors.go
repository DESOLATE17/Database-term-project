package models

import "errors"

var (
	Conflict      = errors.New("Conflict")
	NotFound      = errors.New("NotFound")
	InternalError = errors.New("InternalError")
)
