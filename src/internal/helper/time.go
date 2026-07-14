package helper

import (
	"fmt"
	"math"

	"github.com/halsten-dev/lokyn"
)

func HoursDecFormat(hours float64) string {
	var formatted string

	total := int(math.Round(hours * 60))
	h, m := total/60, total%60

	if h == 0 && m == 0 {
		m = 1
	}

	if h > 0 && m == 0 {
		formatted = fmt.Sprintf("%d"+lokyn.L("h"), h)
	} else if m > 0 && h == 0 {
		formatted = fmt.Sprintf("%d"+lokyn.L("m"), m)
	} else {
		formatted = fmt.Sprintf("%d"+lokyn.L("h")+" %d"+lokyn.L("m"), h, m)
	}

	return formatted
}
