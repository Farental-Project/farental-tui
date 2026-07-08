package request

import (
	"farental/core/data/api"
	"fmt"

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

func CraftGetMaxCraftable(recipeID uint) *resty.Request {
	return get("/craft/maxCraftableAmount").
		SetResult(api.MaxCraftableAmount{}).
		SetQueryParam("recipeID", fmt.Sprintf("%d", recipeID))
}
