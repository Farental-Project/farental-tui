package internal

import (
	"farental/core/config"
	"farental/internal/contentmanager"
	"github.com/go-resty/resty/v2"
)

type AppCtx struct {
	Client         *resty.Client
	Config         *config.Config
	ContentManager *contentmanager.Manager
}

func NewAppCtx() *AppCtx {
	return &AppCtx{
		Config:         config.NewConfig(),
		Client:         resty.New(),
		ContentManager: contentmanager.New(),
	}
}
