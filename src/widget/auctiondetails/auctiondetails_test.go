package auctiondetails

import (
	"farental/core/data/api"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
)

func testAuction(directBuy int) api.AuctionResponse {
	return api.AuctionResponse{
		CurrentBid:        120,
		DirectBuyPrice:    directBuy,
		CurrentBidderName: "Jane Roe",
		SellerName:        "John Doe",
		EndTimestamp:      time.Now().Add(2 * time.Hour),
	}
}

// The box must never render taller than the height it reports, or the dialog
// layout that trusts that number overflows.
func TestRenderedHeightMatchesReportedHeight(t *testing.T) {
	cases := []struct {
		name      string
		directBuy int
	}{
		{"without direct buy", 0},
		{"with direct buy", 500},
	}

	for _, c := range cases {
		w := New()

		auction := testAuction(c.directBuy)
		w.UpdateData(&auction)

		height := w.GetMinSize().Height

		if got := w.GetPreferredSize().Height; got != height {
			t.Errorf("%s: preferred height = %d, min height = %d; they must "+
				"match so the layout treats the box as fixed height",
				c.name, got, height)
		}

		w.Resize(orvyn.NewSize(41, height))

		if got := lipgloss.Height(w.Render()); got != height {
			t.Errorf("%s: rendered height = %d, reported %d",
				c.name, got, height)
		}
	}
}

// A value too wide for its column must be cut, not wrapped: a wrapped row would
// push the box past its reported height.
func TestLongValueDoesNotWrap(t *testing.T) {
	w := New()

	auction := testAuction(0)
	auction.SellerName = strings.Repeat("Wenceslas", 8)
	w.UpdateData(&auction)

	height := w.GetMinSize().Height

	w.Resize(orvyn.NewSize(41, height))

	if got := lipgloss.Height(w.Render()); got != height {
		t.Errorf("rendered height = %d, reported %d; a long value wrapped "+
			"instead of being truncated", got, height)
	}
}

func TestRenderShowsLabelsAndValues(t *testing.T) {
	w := New()

	auction := testAuction(500)
	w.UpdateData(&auction)

	w.Resize(orvyn.NewSize(41, w.GetMinSize().Height))

	view := w.Render()

	for _, want := range []string{"Auction", "Current bid", "Direct buy", "John Doe"} {
		if !strings.Contains(view, want) {
			t.Errorf("rendered box does not contain %q:\n%s", want, view)
		}
	}
}
