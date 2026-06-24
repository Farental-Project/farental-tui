package request

import (
	"farental/core/data/api"

	"github.com/go-resty/resty/v2"
)

func DataGetAllRace() *resty.Request {
	return get("/data/races").SetResult([]api.RaceResponse{})
}

func DataGetEquipmentSlots() *resty.Request {
	return get("/data/equipmentSlots").SetResult([]api.BasicInfoResponse{})
}
