package context

import (
	"farental/internal/config"
	"farental/internal/contentmanager"
	"github.com/go-resty/resty/v2"
)

var (
	Client         *resty.Client
	Config         *config.Config
	ContentManager *contentmanager.Manager

	CharacterID uint
)

func Init() {
	Config = config.NewConfig()
	Client = resty.New()
	Client.SetBaseURL(Config.BaseURL)
	ContentManager = contentmanager.New()
}
