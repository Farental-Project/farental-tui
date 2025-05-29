package api

import (
	"time"
)

type ChatMessageBody struct {
	Message string
}

type ChatMessageResponse struct {
	Timestamp time.Time
	Name      string
	Message   string
}
