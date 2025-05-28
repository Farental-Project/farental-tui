package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func ActivityGetAvailable() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/activity/getAvailable"
	r.SetResult([]api.ActivityResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func ActivityStart(activityID, durationID uint) *resty.Request {
	r := client.R()
	r.Method = resty.MethodPost
	r.URL = "/activity/start"
	r.SetBody(api.ActivityStartBody{ActivityID: activityID, DurationID: durationID})
	r.SetError(api.ErrorResponse{})

	return r
}
