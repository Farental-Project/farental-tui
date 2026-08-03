package auctionlistitem

import (
	"fmt"
	"time"

	"github.com/halsten-dev/lokyn"
)

// EndsIn renders how long a listing has left. now is a parameter so the
// formatting can be tested without waiting for a clock.
func EndsIn(end, now time.Time) string {
	remaining := end.Sub(now)

	if remaining <= 0 {
		return lokyn.L("ended")
	}

	hours := int(remaining.Hours())

	if hours >= 24 {
		return fmt.Sprintf("%dd %dh", hours/24, hours%24)
	}

	if hours >= 1 {
		return fmt.Sprintf("%dh%02d", hours, int(remaining.Minutes())%60)
	}

	return fmt.Sprintf("%dm", int(remaining.Minutes()))
}
