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

func ScriptGetActive() *resty.Request {
	r := client.R()

	r.Method = resty.MethodGet
	r.URL = "/script/active"
	r.SetResult(api.ScriptBasicResponse{})

	return r
}

func ScriptSetActive(ID []byte) *resty.Request {
	r := client.R()

	r.Method = resty.MethodPost
	r.URL = "/script/setActive"
	r.SetBody(api.UUIDBody{
		ID: ID,
	})
	r.SetResult(api.ScriptSetActiveResponse{})
	r.SetError(api.ErrorResponse{})

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

func ScriptGetDetail(ID []byte) *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/script/detail"
	r.SetQueryParam("scriptID", fmt.Sprintf("%x", ID))
	r.SetResult(api.ScriptResponse{})
	r.SetError(api.ErrorResponse{})

	return r

}

func ScriptDelete(ID []byte) *resty.Request {
	r := client.R()

	r.Method = resty.MethodPost
	r.URL = "/script/delete"
	r.SetBody(api.UUIDBody{
		ID: ID,
	})

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
