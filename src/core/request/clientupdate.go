package request

import (
	"github.com/go-resty/resty/v2"
)

// ManifestGet builds a request for the client update manifest. Unlike the
// endpoints elsewhere in this package, the manifest (and FileDownloadGet,
// below) are served by the marketing website, not the API — a different
// service with no fixed host the way the API's baseurl is, since local
// development and tests point it at different servers than production.
// That is why this goes through getWeb with a caller-built endpoint rather
// than get with a path relative to a fixed base.
func ManifestGet(endpoint, lang string) *resty.Request {
	return getWeb(endpoint).SetQueryParam("lang", lang)
}

// FileDownloadGet builds a streaming GET request for a release binary.
// SetDoNotParseResponse tells resty not to buffer the whole (tens-of-MB)
// response body into memory the way it normally would; the caller reads it
// directly off the network via Response.RawBody() instead. See
// internal/updater/apply.go.
func FileDownloadGet(endpoint string) *resty.Request {
	return getWeb(endpoint).SetDoNotParseResponse(true)
}
