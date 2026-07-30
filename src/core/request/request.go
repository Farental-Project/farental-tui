package request

import (
	"farental/core/data/api"
	"log"

	"github.com/go-resty/resty/v2"
)

var (
	client *resty.Client

	// webClient talks to the marketing website (config.WebURL), not the API
	// (baseurl): a different service, used for the client update manifest
	// and release binary downloads. Kept separate from client because the
	// two hosts have nothing else in common (different auth, different
	// response shapes).
	webClient *resty.Client
)

// Init takes a resty.Client
func Init(c *resty.Client) {
	client = c

	log.Println("Request package successfully initialized")
}

// InitWeb takes the resty.Client bound to the website host, mirroring Init
// above for the API client.
func InitWeb(c *resty.Client) {
	webClient = c

	log.Println("Web request package successfully initialized")
}

// newReq builds a request with the given method and url, wired with the
// standard API error binding. Chain SetResult/SetBody/SetQueryParam as needed.
func newReq(method, url string) *resty.Request {
	r := client.R()
	r.Method = method
	r.URL = url

	return r.SetError(api.ErrorResponse{})
}

// newWebReq builds a request against webClient instead of client. Unlike
// newReq, it does not bind api.ErrorResponse: the website is a different
// service and does not use the API's error body shape, and callers on this
// path (internal/updater) check status codes themselves rather than going
// through helper.SendRequest/ExtractError.
func newWebReq(method, url string) *resty.Request {
	r := webClient.R()
	r.Method = method
	r.URL = url

	return r
}

func get(url string) *resty.Request  { return newReq(resty.MethodGet, url) }
func post(url string) *resty.Request { return newReq(resty.MethodPost, url) }
func put(url string) *resty.Request  { return newReq(resty.MethodPut, url) }
func del(url string) *resty.Request  { return newReq(resty.MethodDelete, url) }

// getWeb builds a GET request against the website host. url is the full
// request URL (scheme+host+path), not a path relative to a fixed base: the
// website host varies (production config.WebURL, local dev, and tests each
// point it at a different server), unlike the API client's fixed baseurl.
func getWeb(url string) *resty.Request { return newWebReq(resty.MethodGet, url) }
