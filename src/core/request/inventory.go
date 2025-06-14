package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func InventoryGetFull() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/inventory/full"
	r.SetResult(api.InventoryResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}
