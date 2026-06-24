package context

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/spf13/viper"
)

var (
	Client *resty.Client

	CharacterID   uint
	CharacterInfo *api.CharacterInfoResponse
	RunningTask   *api.TaskResponse

	// ChatContent is in the context because it needs to stay coherent between gamedashboard and chat model.
	ChatContent       []string
	LastChatTimestamp time.Time
)

func Init() {
	Client = resty.New()
	Client.SetBaseURL(viper.GetString("baseurl"))

	CharacterID = 0
	RunningTask = nil
	ChatContent = make([]string, 0)
}

func Reset() {
	var zeroTime time.Time

	CharacterInfo = nil
	RunningTask = nil
	ChatContent = make([]string, 0)
	LastChatTimestamp = zeroTime
}

func Logout() {
	// Logout - Clear the logintoken in config and clear the client cookie.
	viper.Set("logintoken", "")
	viper.WriteConfig()
	Client.Cookies = make([]*http.Cookie, 0)
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

	res, err := helper.Fetch[[]api.ChatMessageResponse](req)

	if err != nil {
		log.Println(err)
		return
	}

	chatMessages := *res

	if len(chatMessages) == 0 {
		return
	}

	titleStyle := orvyn.GetTheme().Style(theme.TitleStyleID)

	for _, message := range chatMessages {
		chatMessage := fmt.Sprintf("%s %s - %s",
			titleStyle.Render(message.Timestamp.Format(time.TimeOnly)),
			titleStyle.Render(message.Name),
			message.Message)

		ChatContent = append(ChatContent, chatMessage)
	}

	LastChatTimestamp = chatMessages[len(chatMessages)-1].Timestamp
}
