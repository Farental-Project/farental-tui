package request

import (
	api "farental/core/data/api"
	"fmt"
	"github.com/go-resty/resty/v2"
)

func DataGetAllRace() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodGet
	r.URL = fmt.Sprintf("%s/data/races", ctx.Config.BaseURL)
	r.SetResult([]api.DataRaceResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}
