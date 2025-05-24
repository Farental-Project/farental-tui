package request

import (
	"farental/core/data/api"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func ActivityGetAvailable() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodGet
	r.URL = fmt.Sprintf("%s/activity/getAvailable", ctx.Config.BaseURL)
	r.SetResult([]api.ActivityResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func ActivityStart(activityID, durationID uint) *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodPost
	r.URL = fmt.Sprintf("%s/activity/start", ctx.Config.BaseURL)
	r.SetBody(api.ActivityStartBody{ActivityID: activityID, DurationID: durationID})
	r.SetError(api.ErrorResponse{})

	return r
}
