package request

import (
	"farental/core/data/api"

	"github.com/go-resty/resty/v2"
)

func NpcGetAvailable() *resty.Request {
	r := client.R()

	r.Method = resty.MethodGet
	r.URL = "/npc/available"
	r.SetResult([]api.NpcResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func NpcTalkTo(id uint) *resty.Request {
	r := client.R()

	r.Method = resty.MethodPost
	r.URL = "/npc/talkTo"
	r.SetBody(api.IDBody{
		ID: id,
	})
	r.SetResult(api.NpcDialogResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}
