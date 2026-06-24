package request

import (
	"farental/core/data/api"

	"github.com/go-resty/resty/v2"
)

func NpcGetAvailable() *resty.Request {
	return get("/npc/available").SetResult([]api.NpcResponse{})
}

func NpcTalkTo(id uint) *resty.Request {
	return post("/npc/talkTo").SetResult(api.NpcDialogResponse{}).SetBody(api.IDBody{
		ID: id,
	})
}
