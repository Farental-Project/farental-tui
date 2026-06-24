package request

import (
	"farental/core/data/api"

	"github.com/go-resty/resty/v2"
)

func AbilityGetAll() *resty.Request {
	return get("/script/abilities").SetResult([]api.AbilityResponse{})
}

func AbilityGet(code string) *resty.Request {
	return get("/script/ability").SetResult(api.AbilityResponse{}).SetQueryParam("Code", code)
}

func AbilityGetAvailable() *resty.Request {
	return get("/script/availableAbilities").SetResult([]api.AbilityResponse{})
}
