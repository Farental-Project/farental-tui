package request

import (
	"farentalapp/core/data/api"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func CharacterGetAll() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodGet
	r.URL = fmt.Sprintf("%s/character/all", ctx.Config.BaseURL)
	r.SetResult([]api.CharacterBasicInfoResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func CharacterCreate() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodPost
	r.URL = fmt.Sprintf("%s/character/create", ctx.Config.BaseURL)
	r.SetError(api.ErrorResponse{})

	return r
}

func CharacterGetInfo() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodGet
	r.URL = fmt.Sprintf("%s/character/info", ctx.Config.BaseURL)
	r.SetResult(api.CharacterInfoResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func CharacterSetActive(id uint) *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodPut
	r.URL = fmt.Sprintf("%s/character/setActive", ctx.Config.BaseURL)
	r.SetQueryParam("characterID", fmt.Sprint(id))
	r.SetError(api.ErrorResponse{})

	return r
}

func CharacterGetActive() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodGet
	r.URL = fmt.Sprintf("%s/character/getActive", ctx.Config.BaseURL)
	r.SetResult(api.CharacterBasicInfoResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func CharacterGetEventLog() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodGet
	r.URL = fmt.Sprintf("%s/character/eventLog", ctx.Config.BaseURL)
	r.SetResult(api.EventLogResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}
