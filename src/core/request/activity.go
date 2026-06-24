package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func ActivityGetAvailable() *resty.Request {
	return get("/activity/available").SetResult([]api.ActivityResponse{})
}

func ActivityStart(activityID, durationID uint) *resty.Request {
	return post("/activity/start").SetBody(api.ActivityStartBody{ActivityID: activityID, DurationID: durationID})
}
