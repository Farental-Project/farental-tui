package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func TravelGetAvailable() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/travel/getAvailable"
	r.SetResult([]api.TravelResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func TravelStart(travelID uint) *resty.Request {
	r := client.R()
	r.Method = resty.MethodPost
	r.URL = "/travel/start"
	r.SetBody(api.TravelStartBody{
		TravelID: travelID,
	})
	r.SetError(api.ErrorResponse{})

	return r
}
