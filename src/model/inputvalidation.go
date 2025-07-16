package model

import (
	"farental/internal/lang"
	"fmt"
	"regexp"
)

func NumericalValidate(s string) error {
	matched, _ := regexp.MatchString(`^\d*$`, s)
	if !matched {
		return fmt.Errorf(lang.L("Only numbers are allowed"))
	}
	return nil
}
