package request

import (
	"farental/core/data/api"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func ScriptGetRuleTypes() *resty.Request {
	return get("/script/ruletypes").SetResult([]api.ScriptRuleTypeResponse{})
}

func ScriptGetRuleType(code string) *resty.Request {
	return get("/script/ruleType").SetResult(api.ScriptRuleTypeResponse{}).SetQueryParam("Code", code)
}

func ScriptGetRuleTypeParamStruct(ruleTypeCode string) *resty.Request {
	return get("/script/ruleTypeParamStruct").SetResult([]api.ScriptRuleTypeStructParam{}).SetQueryParam("Code", ruleTypeCode)
}

func ScriptGetCount() *resty.Request {
	return get("/script/count").SetResult(api.ScriptCountResponse{})
}

func ScriptGetActive() *resty.Request {
	return get("/script/active").SetResult(api.ScriptBasicResponse{})
}

func ScriptSetActive(ID []byte) *resty.Request {
	return post("/script/setActive").SetResult(api.ScriptSetActiveResponse{}).SetBody(api.UUIDBody{
		ID: ID,
	})
}

func ScriptGetOwn() *resty.Request {
	return get("/script/own").SetResult([]api.ScriptBasicResponse{})
}

func ScriptGetPrivate() *resty.Request {
	return get("/script/private").SetResult([]api.ScriptBasicResponse{})
}

func ScriptGetPublic() *resty.Request {
	return get("/script/public").SetResult([]api.ScriptBasicResponse{})
}

func ScriptGetDetail(ID []byte) *resty.Request {
	return get("/script/detail").SetResult(api.ScriptResponse{}).SetQueryParam("scriptID", fmt.Sprintf("%x", ID))
}

func ScriptDelete(ID []byte) *resty.Request {
	return post("/script/delete").SetBody(api.UUIDBody{
		ID: ID,
	})
}

func ScriptSave(script *api.ScriptBody) *resty.Request {
	return post("/script/save").SetResult(api.UUIDResponse{}).SetBody(script)
}
