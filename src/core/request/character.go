package request

import (
	"farental/core/data/api"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func CharacterGetAll() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/character/all"
	r.SetResult([]api.CharacterBasicResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func CharacterCreate() *resty.Request {
	r := client.R()
	r.Method = resty.MethodPost
	r.URL = "/character/create"
	r.SetError(api.ErrorResponse{})

	return r
}

func CharacterGetInfo() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/character/info"
	r.SetResult(api.CharacterInfoResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func CharacterSetActive(id uint) *resty.Request {
	r := client.R()
	r.Method = resty.MethodPut
	r.URL = "/character/setActive"
	r.SetQueryParam("characterID", fmt.Sprint(id))
	r.SetError(api.ErrorResponse{})

	return r
}

func CharacterGetActive() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/character/getActive"
	r.SetResult(api.CharacterBasicResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func CharacterGetEventLog() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/character/eventLog"
	r.SetResult(api.EventLogResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}
