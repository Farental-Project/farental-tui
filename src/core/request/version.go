package request

import (
	"farental/core/data/api"

	"github.com/go-resty/resty/v2"
)

func VersionGet() *resty.Request {
	r := client.R()

	r.Method = resty.MethodGet
	r.URL = "/version"
	r.SetResult(api.DbVersion{})

	return r
}
