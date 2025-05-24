package request

import (
	"farental/core/data/api"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func TaskGetRunning() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodGet
	r.URL = fmt.Sprintf("%s/task/running", ctx.Config.BaseURL)
	r.SetResult(api.TaskResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func TaskClaim() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodPost
	r.URL = fmt.Sprintf("%s/task/claim", ctx.Config.BaseURL)
	r.SetError(api.ErrorResponse{})

	return r
}
