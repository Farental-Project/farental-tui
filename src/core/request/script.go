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

func ScriptGetRuleType(code string) *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/script/ruleType"
	r.SetQueryParam("Code", code)
	r.SetResult(api.ScriptRuleTypeResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func ScriptGetRuleTypeParamStruct(ruleTypeCode string) *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/script/ruleTypeParamStruct"
	r.SetQueryParam("Code", ruleTypeCode)
	r.SetResult([]api.ScriptRuleTypeStructParam{})
	r.SetError(api.ErrorResponse{})

	return r

}

func ScriptGetCount() *resty.Request {
	r := client.R()

	r.Method = resty.MethodGet
	r.URL = "/script/count"
	r.SetResult(api.ScriptCountResponse{})

	return r
}

func ScriptGetOwn() *resty.Request {
	r := client.R()

	r.Method = resty.MethodGet
	r.URL = "/script/own"
	r.SetResult([]api.ScriptBasicResponse{})

	return r
}

func ScriptGetPrivate() *resty.Request {
	r := client.R()

	r.Method = resty.MethodGet
	r.URL = "/script/private"
	r.SetResult([]api.ScriptBasicResponse{})

	return r
}

func ScriptGetPublic() *resty.Request {
	r := client.R()

	r.Method = resty.MethodGet
	r.URL = "/script/public"
	r.SetResult([]api.ScriptBasicResponse{})

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

func ScriptSave(script *api.ScriptBody) *resty.Request {
	r := client.R()

	r.Method = resty.MethodPost
	r.URL = "/script/save"
	r.SetBody(script)
	r.SetError(api.ErrorResponse{})

	return r
}
