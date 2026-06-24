package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func CraftGetAvailable() *resty.Request {
	return get("/craft/available").SetResult([]api.RecipeResponse{})
}

func CraftStart(craftID uint, amount int) *resty.Request {
	return post("/craft/start").SetBody(api.CraftStartBody{
		RecipeID: craftID,
		Amount:   amount,
	})
}
