package request

import (
	"farental/core/data/api"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func CharacterGetAll() *resty.Request {
	return get("/character/all").SetResult([]api.CharacterBasicResponse{})
}

func CharacterCreate() *resty.Request {
	return post("/character/create")
}

func CharacterGetInfo() *resty.Request {
	return get("/character/info").SetResult(api.CharacterInfoResponse{})
}

func CharacterInspect(id uint) *resty.Request {
	return get("/character/inspect").
		SetQueryParam("characterID", fmt.Sprint(id)).
		SetResult(api.CharacterInspectResponse{})
}

func CharacterSetActive(id uint) *resty.Request {
	return put("/character/setActive").SetQueryParam("characterID", fmt.Sprint(id))
}

func CharacterGetActive() *resty.Request {
	return get("/character/active").SetResult(api.CharacterBasicResponse{})
}

func CharacterGetEventLog() *resty.Request {
	return get("/character/eventLog").SetResult(api.EventLogResponse{})
}

func CharacterGetCurrencyAmount(currencyCode api.CurrencyCode) *resty.Request {
	return get("/character/currencyAmount").
		SetResult(api.CurrencyResponse{}).
		SetQueryParam("currencyCode", fmt.Sprint(currencyCode))
}

func CharacterHaveBankAccount() *resty.Request {
	return get("/character/haveBankAccount")
}
