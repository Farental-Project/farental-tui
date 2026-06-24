package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func TaskGetRunning() *resty.Request {
	return get("/task/running").SetResult(api.TaskResponse{})
}

func TaskClaim() *resty.Request {
	return post("/task/claim")
}
