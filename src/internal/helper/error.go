package helper

import (
	"errors"
	"farental/core/data/api"
	"github.com/go-resty/resty/v2"
	"github.com/halsten-dev/lokyn"
	"strings"
)

func ExtractError(resp *resty.Response) error {
	var b strings.Builder

	if resp.StatusCode() == 200 || resp.StatusCode() == 201 || resp.StatusCode() == 302 {
		return nil
	}

	if resp.Error() != nil {
		errorResp, ok := resp.Error().(*api.ErrorResponse)

		if !ok {
			return nil
		}

		for i, err := range errorResp.Errors {
			if i > 0 {
				b.WriteString("\n")
			}

			b.WriteString(err.Message)

			if i >= 2 {
				break
			}
		}

		return errors.New(b.String())
	}

	return nil
}

func ConnectionError() error {
	return errors.New(lokyn.L("Cannot connect to Farental server"))
}
