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

// The dialog asks the box to stretch by setting a preferred size. Without that,
// the box must keep reporting its own row height, which is what makes the
// layout treat it as fixed.
func TestPreferredSizeHonoursAnExplicitStretch(t *testing.T) {
	w := New()

	auction := testAuction(500)
	w.UpdateData(&auction)

	if got, want := w.GetPreferredSize().Height, w.GetMinSize().Height; got != want {
		t.Errorf("unstretched preferred height = %d, want %d (the min height)",
			got, want)
	}

	w.SetPreferredSize(orvyn.NewSize(1, 200))

	if got := w.GetPreferredSize().Height; got != 200 {
		t.Errorf("stretched preferred height = %d, want 200", got)
	}

	if got, want := w.GetMinSize().Height, 9; got != want {
		t.Errorf("min height = %d, want %d; the stretch must not move the "+
			"minimum, which is what the box actually needs", got, want)
	}
}

// The layout gives both boxes in the dialog the same height, so each must render
// at exactly that height for their outlines to align. iteminspect clips to its
// allocation; this box must too. Rendering taller - which lipgloss Height would
// happily allow, since it pads but never truncates - is what made the two
// outlines diverge on short terminals.
func TestRenderClipsToAStarvedAllocation(t *testing.T) {
	w := New()

	auction := testAuction(500)
	w.UpdateData(&auction)

	natural := w.GetMinSize().Height

	for _, allocated := range []int{natural, natural - 2, natural - 4} {
		w.Resize(orvyn.NewSize(41, allocated))

		if got := lipgloss.Height(w.Render()); got != allocated {
			t.Errorf("allocated %d, rendered %d", allocated, got)
		}
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
