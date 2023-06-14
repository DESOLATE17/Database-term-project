package models

// easyjson -all ./internal/models/errorResponse.go

type ErrorResponse struct {
	Message string `json:"message"`
}
