package request

import (
	"farental/core/data/api"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func FightGetFinished() *resty.Request {
	r := client.R()

	r.Method = resty.MethodGet
	r.URL = "/fight/finished"
	r.SetResult([]api.FightResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func FightGetLog(fightID uint) *resty.Request {
	r := client.R()

	r.Method = resty.MethodGet
	r.URL = "/fight/eventLog"
	r.SetQueryParam("fightID", fmt.Sprintf("%d", fightID))
	r.SetResult(api.EventLogResponse{})

	return r
}

func FightGetAvailable() *resty.Request {
	r := client.R()

	r.Method = resty.MethodGet
	r.URL = "/fight/available"
	r.SetResult([]api.FightCompositionResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func FightStart(fightCompoID uint) *resty.Request {
	r := client.R()

	r.Method = resty.MethodPost
	r.URL = "/fight/start"
	r.SetBody(api.IDBody{ID: fightCompoID})
	r.SetError(api.ErrorResponse{})

	return r
}
