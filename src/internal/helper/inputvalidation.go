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
		return fmt.Errorf("%s", lokyn.L("Only numbers are allowed"))
	}
	return nil
}

// SignedNumericalValidate allows a leading minus, which NumericalValidate does
// not: the auction filter's minimum stat has to express penalties.
func SignedNumericalValidate(s string) error {
	matched, _ := regexp.MatchString(`^-?\d*$`, s)
	if !matched {
		return fmt.Errorf("%s", lokyn.L("Only numbers are allowed"))
	}
	return nil
}

func EmailIsValid(s string) bool {
	_, err := mail.ParseAddress(s)

	return err == nil
}
