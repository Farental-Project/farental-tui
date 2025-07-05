package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func LocationGetCharacters() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/location/characters"
	r.SetResult([]api.CharacterBasicWithActivityResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func LocationTavernSleep() *resty.Request {
	r := client.R()
	r.Method = resty.MethodPost
	r.URL = "/location/tavern/sleep"
	r.SetError(api.ErrorResponse{})

	return r
}
