package helper

import (
	"farental/internal/lang"
	"fmt"
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
		formatted = fmt.Sprintf("%d"+lang.L("h"), h)
	} else if m > 0 && h == 0 {
		formatted = fmt.Sprintf("%d"+lang.L("m"), m)
	} else {
		formatted = fmt.Sprintf("%d"+lang.L("h")+" %d"+lang.L("m"), h, m)
	}

	return formatted
}
