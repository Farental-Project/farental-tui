package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func ChatGetMessages() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/chat/messages"
	r.SetResult([]api.ChatMessageResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func ChatSendMessage() *resty.Request {
	r := client.R()
	r.Method = resty.MethodPost
	r.URL = "/chat/send"
	r.SetError(api.ErrorResponse{})

	return r
}
