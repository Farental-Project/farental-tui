package request

import (
	"farental/core/data/api"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func TaskGetRunning() *resty.Request {
	return get("/task/running").SetResult(api.TaskResponse{})
}

func TaskInspect(id uint) *resty.Request {
	return get("/task/inspect").
		SetQueryParam("characterID", fmt.Sprint(id)).
		SetResult(api.TaskResponse{})
}

func TaskClaim() *resty.Request {
	return post("/task/claim")
}
