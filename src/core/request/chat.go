package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func ChatGetMessages() *resty.Request {
	return get("/chat/messages").SetResult([]api.ChatMessageResponse{})
}

func ChatSendMessage() *resty.Request {
	return post("/chat/send")
}
