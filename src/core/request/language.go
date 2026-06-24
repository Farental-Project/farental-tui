package request

import (
	"farental/core/data/api"

	"github.com/go-resty/resty/v2"
)

func LangGetAll() *resty.Request {
	return get("/languages").SetResult([]api.LanguageResponse{})
}
