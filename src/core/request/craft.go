package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func CraftGetAvailable() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/craft/getAvailable"
	r.SetResult([]api.RecipeResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func CraftStart(craftID uint, amount int) *resty.Request {
	r := client.R()
	r.Method = resty.MethodPost
	r.URL = "/craft/start"
	r.SetBody(api.CraftStartBody{
		RecipeID: craftID,
		Amount:   amount,
	})
	r.SetError(api.ErrorResponse{})

	return r
}
