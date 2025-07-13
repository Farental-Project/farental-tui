package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func DataGetAllRace() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/data/races"
	r.SetResult([]api.RaceResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func DataGetEquipmentSlots() *resty.Request {
	r := client.R()

	r.Method = resty.MethodGet
	r.URL = "/data/equipmentSlots"
	r.SetResult([]api.BasicInfoResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}
