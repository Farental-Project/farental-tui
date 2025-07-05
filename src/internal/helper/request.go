package helper

import (
	"github.com/go-resty/resty/v2"
)

func SendRequest(req *resty.Request) (*resty.Response, error) {
	resp, err := req.Send()

	if err != nil {
		err = ConnectionError()
		return nil, err
	}

	err = ExtractError(resp)

	if err != nil {
		return nil, err
	}

	return resp, nil
}
