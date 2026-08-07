package auctionfilter

import (
	"farental/core/data/api"
	"farental/internal/keybind"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

// New reaches lokyn through Reset, which dereferences a localizer that only
// SetLanguage creates. keybind.Up/Down are package vars keybind.Init sets;
// without it they are zero bindings that match nothing.
func newTestWidget(t *testing.T) *Widget {
	t.Helper()

	orvyn.Init()
	lokyn.Init()
	lokyn.SetLanguage("en")
	keybind.Init()

	w := New()
	w.Init()

	return w
}

// runeMsg builds the tea.KeyMsg a real keypress of r would produce.
func runeMsg(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
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

// The screen stands its letter keybinds down on this, so it must be false
// whenever the panel itself lacks focus, not just when its cursor moves.
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

// j and k are ordinary letters in item names (jerkin, cloak); the box must
// keep every one instead of losing them to focus movement.
func TestTypingIntoSearchKeepsEveryLetter(t *testing.T) {
	w := newTestWidget(t)

	screenFocus := orvyn.NewFocusManager()
	screenFocus.SetWidgets([]orvyn.Focusable{w})
	screenFocus.FocusFirst()

	for _, r := range "jerkin" {
		w.Update(runeMsg(r))
	}

	if got := w.GetSearch(); got != "jerkin" {
		t.Errorf("GetSearch after typing = %q, want %q", got, "jerkin")
	}

	if !w.SearchFocused() {
		t.Errorf("SearchFocused = false after typing, want true: the cursor must stay on the box")
	}
}

// The fix must not cost the panel its vim navigation: j/k still have to move
// the cursor once it is off the search box.
func TestVimAliasesMoveFocusOffTheSearchBox(t *testing.T) {
	w := newTestWidget(t)

	screenFocus := orvyn.NewFocusManager()
	screenFocus.SetWidgets([]orvyn.Focusable{w})
	screenFocus.FocusFirst()

	w.focusManager.NextFocus()
	start := w.focusManager.TabIndex()

	w.Update(runeMsg('j'))

	if got := w.focusManager.TabIndex(); got != start+1 {
		t.Fatalf("TabIndex after j = %d, want %d", got, start+1)
	}

	w.Update(runeMsg('k'))

	if got := w.focusManager.TabIndex(); got != start {
		t.Errorf("TabIndex after k = %d, want %d", got, start)
	}
}
