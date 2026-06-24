package request

import (
	"farental/core/data/api"

	"github.com/go-resty/resty/v2"
)

func VersionGet() *resty.Request {
	return get("/version").SetResult(api.DbVersion{})
}
