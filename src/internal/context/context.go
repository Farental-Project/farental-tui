package context

import (
	"farental/core/data/api"
	"farental/internal/contentmanager"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

var (
	Client         *resty.Client
	ContentManager *contentmanager.Manager

	CharacterID uint
	RunningTask *api.TaskResponse
)

func Init() {
	Client = resty.New()
	Client.SetBaseURL(viper.GetString("baseurl"))
	ContentManager = contentmanager.New()

	CharacterID = 0
	RunningTask = nil
}
