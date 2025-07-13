package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func LangGetAll() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/languages"
	r.SetResult([]api.LanguageResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}
