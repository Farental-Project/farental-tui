package helper

import (
	"fmt"
	"net/mail"
	"regexp"

	"github.com/halsten-dev/lokyn"
)

func NumericalValidate(s string) error {
	matched, _ := regexp.MatchString(`^\d*$`, s)
	if !matched {
		return fmt.Errorf(lokyn.L("Only numbers are allowed"))
	}
	return nil
}

func EmailIsValid(s string) bool {
	_, err := mail.ParseAddress(s)

	return err == nil
}
