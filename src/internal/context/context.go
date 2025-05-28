package context

import (
	"farental/core/config"
	"farental/internal/contentmanager"
	"github.com/go-resty/resty/v2"
)

var (
	Client         *resty.Client
	Config         *config.Config
	ContentManager *contentmanager.Manager
)

func Init() {
	Config = config.NewConfig()
	Client = resty.New()
	Client.SetBaseURL(Config.BaseURL)
	ContentManager = contentmanager.New()
}
