package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func ScriptGetPrivate() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/script/private"
	r.SetResult([]api.ScriptBasicResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}
