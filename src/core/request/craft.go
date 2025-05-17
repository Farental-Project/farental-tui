package request

import (
	"farentalapp/core/data/api"
	"fmt"
	"github.com/go-resty/resty/v2"
)

func CraftGetAvailable() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodGet
	r.URL = fmt.Sprintf("%s/craft/getAvailable", ctx.Config.BaseURL)
	r.SetResult([]api.RecipeResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func CraftStart(craftID, amount uint) *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodPost
	r.URL = fmt.Sprintf("%s/craft/start", ctx.Config.BaseURL)
	r.SetBody(api.CraftStartBody{
		RecipeID: craftID,
		Amount:   amount,
	})
	r.SetError(api.ErrorResponse{})

	return r
}
