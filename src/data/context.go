package data

import (
	"farental/core/config"
	"github.com/go-resty/resty/v2"
)

type AppCtx struct {
	Client *resty.Client
	Config *config.Config
}

func NewAppCtx() *AppCtx {
	return &AppCtx{
		Config: config.NewConfig(),
	}
}
