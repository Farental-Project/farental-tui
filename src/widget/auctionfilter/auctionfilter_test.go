package auctionfilter

import (
	"farental/core/data/api"
	"testing"

	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

// New reaches lokyn through Reset, which dereferences a localizer that only
// SetLanguage creates.
func newTestWidget(t *testing.T) *Widget {
	t.Helper()

	orvyn.Init()
	lokyn.Init()
	lokyn.SetLanguage("en")

	w := New()
	w.Init()

	return w
}

// The search box narrows rows that are already on screen, so its text must
// never reach the query the request layer sends.
func TestSearchNeverReachesTheQuery(t *testing.T) {
	w := newTestWidget(t)

	w.tiSearch.SetValue("iron sword")

	if got := w.GetSearch(); got != "iron sword" {
		t.Errorf("GetSearch = %q, want %q", got, "iron sword")
	}

	if got := w.GetFilter(); got != (api.AuctionFilter{}) {
		t.Errorf("GetFilter = %+v, want the zero filter", got)
	}
}

func TestResetClearsTheSearch(t *testing.T) {
	w := newTestWidget(t)

	w.tiSearch.SetValue("iron sword")

	w.Reset()

	if got := w.GetSearch(); got != "" {
		t.Errorf("GetSearch after Reset = %q, want empty", got)
	}
}

// The screen stands its single-letter keybinds down on this, so it has to be
// false whenever the panel does not hold the screen's focus - the panel keeps
// its own cursor where it was while the auction list is being used.
func TestSearchFocusedNeedsThePanelToHoldFocus(t *testing.T) {
	w := newTestWidget(t)

	screenFocus := orvyn.NewFocusManager()
	screenFocus.SetWidgets([]orvyn.Focusable{w})
	screenFocus.FocusFirst()

	if !w.SearchFocused() {
		t.Errorf("SearchFocused = false with the panel focused on its first control, want true")
	}

	screenFocus.BlurCurrent()

	if w.SearchFocused() {
		t.Errorf("SearchFocused = true with the panel blurred, want false")
	}
}

// Moving off the box is what gives the screen its letter keys back.
func TestSearchFocusedFollowsThePanelCursor(t *testing.T) {
	w := newTestWidget(t)

	screenFocus := orvyn.NewFocusManager()
	screenFocus.SetWidgets([]orvyn.Focusable{w})
	screenFocus.FocusFirst()

	w.focusManager.NextFocus()

	if w.SearchFocused() {
		t.Errorf("SearchFocused = true with the panel cursor moved off the box, want false")
	}
}
