package request

import (
	"farentalapp/core/data/api"
	"fmt"
	"github.com/go-resty/resty/v2"
)

func TravelGetAvailable() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodGet
	r.URL = fmt.Sprintf("%s/travel/getAvailable", ctx.Config.BaseURL)
	r.SetResult([]api.TravelResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func TravelStart(travelID uint) *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodPost
	r.URL = fmt.Sprintf("%s/travel/start", ctx.Config.BaseURL)
	r.SetBody(api.TravelStartBody{
		TravelID: travelID,
	})
	r.SetError(api.ErrorResponse{})

	return r
}
