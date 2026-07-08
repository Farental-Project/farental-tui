package api

import (
	"time"
)

type FeedbackKind string

const (
	KindQuestion  FeedbackKind = "Question"
	KindFeedback  FeedbackKind = "Feedback"
	KindBugReport FeedbackKind = "Bug report"
)

func (f FeedbackKind) IsValid() bool {
	switch f {
	case KindQuestion, KindFeedback, KindBugReport:
		return true
	}
	return false
}

type PlatformName string

const (
	PlatformUndefined PlatformName = ""
	PlatformWindows   PlatformName = "Windows"
	PlatformLinux     PlatformName = "Linux"
	PlatformMacOS     PlatformName = "MacOS"
)

func (p PlatformName) IsValid() bool {
	switch p {
	case PlatformUndefined, PlatformWindows, PlatformLinux, PlatformMacOS:
		return true
	}

	return false
}

type UserResponse struct {
	ID                    uint
	Username              string
	Email                 string
	FirstName             string
	LastName              string
	LanguageCode          string
	WantsNewsletter       bool
	IsAdmin               bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
	LastLoginAt           *time.Time
	AllowedCharacterCount int
}

type UserSettingsBody struct {
	WantsNewsletter bool
	LanguageCode    string
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

type SendFeedbackBody struct {
	Subject  string
	Message  string
	Kind     FeedbackKind
	Platform PlatformName
}

type AuthSuccessResponse struct {
	Data string
}
