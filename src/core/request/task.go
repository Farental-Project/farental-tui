package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func TaskGetRunning() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/task/running"
	r.SetResult(api.TaskResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func TaskClaim() *resty.Request {
	r := client.R()
	r.Method = resty.MethodPost
	r.URL = "/task/claim"
	r.SetError(api.ErrorResponse{})

	return r
}
