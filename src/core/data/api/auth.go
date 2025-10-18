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

type AuthSignUpBody struct {
	Username        string
	Email           string
	Password        string
	PasswordConfirm string
}

type AuthLoginBody struct {
	Email    string
	Password string
}

type AuthSuccessResponse struct {
	Data string
}
