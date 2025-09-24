package request

import (
	"farental/core/data/api"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func ScriptGetRuleTypes() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/script/ruletypes"
	r.SetResult([]api.ScriptRuleTypeResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func ScriptGetPrivate() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/script/private"
	r.SetResult([]api.ScriptBasicResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func ScriptGetDetail(ID uint) *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/script/detail"
	r.SetQueryParam("scriptID", fmt.Sprintf("%d", ID))
	r.SetResult(api.ScriptResponse{})
	r.SetError(api.ErrorResponse{})

	return r

}
