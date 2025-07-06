package request

import (
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
)

func MailGetAll() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/mail/all"
	r.SetResult([]api.MailBasicResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}
