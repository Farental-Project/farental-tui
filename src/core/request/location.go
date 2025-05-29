package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func LocationGetCharacters() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/location/getCharacters"
	r.SetResult([]api.CharacterBasicWithActivityResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}
