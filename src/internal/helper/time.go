package helper

import (
	"fmt"
	"github.com/halsten-dev/lokyn"
	"time"
)

func HoursDecFormat(hours float64) string {
	var formatted string

	duration := time.Duration(hours * float64(time.Hour))
	h := int(duration.Hours())
	m := int(duration.Minutes()) % 60

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
