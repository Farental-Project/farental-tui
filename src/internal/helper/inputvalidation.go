package helper

import (
	"fmt"
	"github.com/halsten-dev/lokyn"
	"regexp"
)

func NumericalValidate(s string) error {
	matched, _ := regexp.MatchString(`^\d*$`, s)
	if !matched {
		return fmt.Errorf(lokyn.L("Only numbers are allowed"))
	}
	return nil
}
