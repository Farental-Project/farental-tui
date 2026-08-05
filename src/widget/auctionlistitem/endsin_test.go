package auctionlistitem

import (
	"os"
	"testing"
	"time"

	"github.com/halsten-dev/lokyn"
)

func TestMain(m *testing.M) {
	lokyn.Init()
	lokyn.SetLanguage("en")

	os.Exit(m.Run())
}

func TestEndsIn(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		end  time.Time
		want string
	}{
		{"hours and minutes", now.Add(4*time.Hour + 2*time.Minute), "4h02"},
		{"whole hours", now.Add(12 * time.Hour), "12h00"},
		{"more than a day", now.Add(27 * time.Hour), "1d 3h"},
		{"under an hour", now.Add(45 * time.Minute), "45m"},
		{"already over", now.Add(-time.Minute), "ended"},
	}

	for _, c := range cases {
		if got := EndsIn(c.end, now); got != c.want {
			t.Errorf("%s: EndsIn = %q, want %q", c.name, got, c.want)
		}
	}
}
