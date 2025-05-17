package request

import (
	"farentalapp/core/data/api"
	"fmt"
	"github.com/go-resty/resty/v2"
)

func FightGetFinished() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodGet
	r.URL = fmt.Sprintf("%s/fight/getFinished", ctx.Config.BaseURL)
	r.SetResult([]api.FightResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func FightGetAvailable() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodGet
	r.URL = fmt.Sprintf("%s/fight/getAvailable", ctx.Config.BaseURL)
	r.SetResult([]api.FightCompositionResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func FightStart(fightCompoID uint) *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodPost
	r.URL = fmt.Sprintf("%s/fight/start", ctx.Config.BaseURL)
	r.SetBody(api.IDBody{ID: fightCompoID})
	r.SetError(api.ErrorResponse{})

	return r
}
