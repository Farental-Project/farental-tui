package context

import (
	"farental/core/data/api"
	"farental/internal/config"
	"farental/internal/contentmanager"
	"github.com/go-resty/resty/v2"
)

var (
	Client         *resty.Client
	Config         *config.Config
	ContentManager *contentmanager.Manager

	CharacterID uint
	RunningTask *api.TaskResponse
)

func Init() {
	Config = config.NewConfig()
	Client = resty.New()
	Client.SetBaseURL(Config.BaseURL)
	ContentManager = contentmanager.New()

	CharacterID = 0
	RunningTask = nil
}
