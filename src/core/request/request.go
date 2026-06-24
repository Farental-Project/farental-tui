package request

import (
	"farental/core/data/api"
	"log"

	"github.com/go-resty/resty/v2"
)

var client *resty.Client

// Init takes a resty.Client
func Init(c *resty.Client) {
	client = c

	log.Println("Request package successfully initialized")
}

// newReq builds a request with the given method and url, wired with the
// standard API error binding. Chain SetResult/SetBody/SetQueryParam as needed.
func newReq(method, url string) *resty.Request {
	r := client.R()
	r.Method = method
	r.URL = url

	return r.SetError(api.ErrorResponse{})
}

func get(url string) *resty.Request  { return newReq(resty.MethodGet, url) }
func post(url string) *resty.Request { return newReq(resty.MethodPost, url) }
func put(url string) *resty.Request  { return newReq(resty.MethodPut, url) }
func del(url string) *resty.Request  { return newReq(resty.MethodDelete, url) }
