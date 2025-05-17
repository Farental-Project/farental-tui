package request

import (
	api "farentalapp/core/data/api"
	"fmt"
	"github.com/go-resty/resty/v2"
)

func LangGetAll() *resty.Request {
	r := ctx.Client.R()
	r.Method = resty.MethodGet
	r.URL = fmt.Sprintf("%s/languages", ctx.Config.BaseURL)
	r.SetResult([]api.LanguageResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}
