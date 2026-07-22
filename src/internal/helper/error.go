package helper

import (
	"errors"
	"farental/core/data/api"
	"farental/internal/session"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/halsten-dev/lokyn"
)

func ExtractError(resp *resty.Response) error {

	if resp.StatusCode() == 200 || resp.StatusCode() == 201 || resp.StatusCode() == 302 {
		return nil
	}

	err := extractErrorMessage(resp)

	// A 401 without any error message is the JWT middleware rejecting the
	// token (expired or revoked); business 401 errors always carry a message.
	if resp.StatusCode() == http.StatusUnauthorized && (err == nil || err.Error() == "") {
		session.Expire()
		return errors.New(lokyn.L(session.ExpiredMessage))
	}

	return err
}

func extractErrorMessage(resp *resty.Response) error {
	var b strings.Builder

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
