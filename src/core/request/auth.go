package request

import (
	api "farentalapp/core/data/api"
	"fmt"
	"github.com/go-resty/resty/v2"
)

func Login() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodPost
	r.URL = fmt.Sprintf("%s/auth/login", ctx.Config.BaseURL)
	r.SetResult(api.AuthSuccessResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func AuthInfo() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodGet
	r.URL = fmt.Sprintf("%s/auth/info", ctx.Config.BaseURL)
	r.SetResult(api.DataResponse[api.UserResponse]{})

	return r
}
