package api

import (
	"time"
)

type UserResponse struct {
	ID                    uint
	Username              string
	Email                 string
	FirstName             string
	LastName              string
	WantsNewsletter       bool
	IsAdmin               bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
	LastLoginAt           *time.Time
	AllowedCharacterCount int
}

type AuthLoginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthSuccessResponse struct {
	Data string
}
