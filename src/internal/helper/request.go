package helper

import (
	"errors"

	"github.com/go-resty/resty/v2"
	"github.com/halsten-dev/lokyn"
)

func SendRequest(req *resty.Request) (*resty.Response, error) {
	resp, err := req.Send()

	if err != nil {
		err = ConnectionError()
		return nil, err
	}

	err = ExtractError(resp)

	if err != nil {
		if err.Error() != "" {
			return nil, err
		}
	}

	return resp, nil
}

// Fetch sends the request and returns its typed result. The result is asserted
// to *T (the type registered via SetResult); a mismatch returns an error rather
// than panicking. On any send/API error the result is nil and the error is set.
func Fetch[T any](req *resty.Request) (*T, error) {
	resp, err := SendRequest(req)

	if err != nil {
		return nil, err
	}

	result, ok := resp.Result().(*T)

	if !ok {
		return nil, errors.New(lokyn.L("Unexpected server response"))
	}

	return result, nil
}
