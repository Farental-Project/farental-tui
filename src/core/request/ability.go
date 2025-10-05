package request

import (
	"farental/core/data/api"

	"github.com/go-resty/resty/v2"
)

func AbilityGetAll() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/script/abilities"
	r.SetResult([]api.AbilityResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func AbilityGet(code string) *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/script/ability"
	r.SetQueryParam("Code", code)
	r.SetResult(api.AbilityResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func AbilityGetAvailable() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/script/availableAbilities"
	r.SetResult([]api.AbilityResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}
