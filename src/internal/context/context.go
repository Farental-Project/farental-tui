package context

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/contentmanager"
	"farental/style"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"log"
	"time"
)

var (
	Client         *resty.Client
	ContentManager *contentmanager.Manager

	CharacterID   uint
	CharacterInfo *api.CharacterInfoResponse
	RunningTask   *api.TaskResponse

	// ChatContent is in the context because it need to stay coherent between gamedashboard and chat model.
	ChatContent       []string
	LastChatTimestamp time.Time
)

func Init() {
	Client = resty.New()
	Client.SetBaseURL(viper.GetString("baseurl"))
	ContentManager = contentmanager.New()

	CharacterID = 0
	RunningTask = nil
	ChatContent = make([]string, 0)
}

func UpdateChat() {
	var req *resty.Request
	var queryParam string

	req = request.ChatGetMessages()

	queryParam = ""
	length := len(ChatContent)

	if length > 0 {
		queryParam = LastChatTimestamp.Format(time.DateTime)
	}

	req.SetQueryParam("lastTimestamp", queryParam)

	resp, err := req.Send()

	if err != nil {
		log.Println(err)
		return
	}

	chatMessages := *resp.Result().(*[]api.ChatMessageResponse)

	if len(chatMessages) == 0 {
		return
	}

	for _, message := range chatMessages {
		chatMessage := fmt.Sprintf("%s %s - %s",
			style.TitleStyle.Render(message.Timestamp.Format(time.TimeOnly)),
			style.TitleStyle.Render(message.Name),
			message.Message)

		ChatContent = append(ChatContent, chatMessage)
	}

	LastChatTimestamp = chatMessages[len(chatMessages)-1].Timestamp

	return
}
