package request

import (
	"farental/core/data/api"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func FightGetFinished() *resty.Request {
	return get("/fight/finished").SetResult([]api.FightResponse{})
}

func FightGetLog(fightID uint) *resty.Request {
	return get("/fight/eventLog").
		SetResult(api.EventLogResponse{}).
		SetQueryParam("fightID", fmt.Sprintf("%d", fightID))
}

func FightGetAvailable() *resty.Request {
	return get("/fight/available").SetResult([]api.FightCompositionResponse{})
}

func FightStart(fightCompoID uint) *resty.Request {
	return post("/fight/start").SetBody(api.IDBody{ID: fightCompoID})
}
