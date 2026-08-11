package simplelogviewer

import (
	ftheme "farental/internal/theme"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
)

func TestMain(m *testing.M) {
	orvyn.Init()
	orvyn.SetTheme(ftheme.NewFarentalDarkTheme())

	os.Exit(m.Run())
}

func lines(n int) []string {
	content := make([]string, 0, n)

	for i := range n {
		content = append(content, fmt.Sprintf("line %d", i))
	}

	return content
}

// Without explicit sizes the viewer is fixed height: it must not claim leftover
// height on its own, because how much room it deserves is the screen's call, not
// something the content can tell it.
func TestDerivedSizesAreFixedHeight(t *testing.T) {
	w := New("Description")
	w.SetAutoScroll(false)
	w.Resize(orvyn.NewSize(40, 20))
	w.SetContent(lines(30))

	minHeight := w.GetMinSize().Height
	prefHeight := w.GetPreferredSize().Height

	if prefHeight != minHeight {
		t.Errorf("preferred height = %d, min height = %d; without explicit sizes "+
			"they must match so the layout treats the viewer as fixed height",
			prefHeight, minHeight)
	}
}

// The regression this guards: the widget reported the BaseRenderable default of
// 1x1, which a vertical layout read as a fixed one-row widget. The border ate
// the row and the viewer rendered an empty box.
func TestMinHeightShowsAContentLine(t *testing.T) {
	for _, title := range []string{"Description", ""} {
		w := New(title)
		w.SetAutoScroll(false)
		w.SetContent([]string{"only line"})

		height := w.GetMinSize().Height

		w.Resize(orvyn.NewSize(40, height))

		view := w.Render()

		if !strings.Contains(view, "only line") {
			t.Errorf("title %q: rendered at its own min height of %d rows, the "+
				"viewer shows no content:\n%s", title, height, view)
		}

		if got := lipgloss.Height(view); got != height {
			t.Errorf("title %q: rendered height = %d, reported a minimum of %d",
				title, got, height)
		}
	}
}

// The screen-level symptom, reproduced on the widget alone: inside a vertical
// layout whose children are fixed height, the viewer was allocated a single row,
// which its border consumed whole, and it rendered an empty box. It stays fixed
// height here - only the screen can hand it more - but it must show content.
func TestVerticalLayoutLeavesTheViewerReadable(t *testing.T) {
	w := New("Description")
	w.SetAutoScroll(false)
	w.SetContent(lines(30))

	l := layout.NewDefinedWidthVerticalLayout(35, 80, orvyn.NewSize(10, 4),
		orvyn.NewSimpleRenderable("Title"),
		w,
	)

	l.Resize(orvyn.NewSize(100, 40))

	if view := l.Render(); !strings.Contains(view, "line 0") {
		t.Errorf("the viewer rendered no content:\n%s", view)
	}
}

// Declaring a preferred height above the minimum is how a screen makes the viewer
// flexible, and the layout must then share its leftover height with it.
func TestExplicitPreferredHeightClaimsLeftoverHeight(t *testing.T) {
	w := New("Description")
	w.SetAutoScroll(false)
	w.SetMinSize(orvyn.NewSize(10, 5))
	w.SetPreferredSize(orvyn.NewSize(30, 35))
	w.SetContent(lines(30))

	l := layout.NewDefinedWidthVerticalLayout(35, 80, orvyn.NewSize(10, 4),
		orvyn.NewSimpleRenderable("Title"),
		w,
	)

	l.Resize(orvyn.NewSize(100, 40))
	l.Render()

	if got := lipgloss.Height(w.Render()); got < 20 {
		t.Errorf("viewer height = %d rows in a 40-row layout; the leftover "+
			"height was not shared with it", got)
	}
}

// Screens pin a viewer with SetMinSize / SetPreferredSize; those must win over
// the sizes the widget derives from its content.
func TestExplicitSizesWin(t *testing.T) {
	w := New("Description")
	w.Resize(orvyn.NewSize(40, 20))
	w.SetContent(lines(30))

	w.SetMinSize(orvyn.NewSize(3, 10))
	w.SetPreferredSize(orvyn.NewSize(4, 12))

	if got := w.GetMinSize(); got != orvyn.NewSize(3, 10) {
		t.Errorf("min size = %v, want the explicit %v", got, orvyn.NewSize(3, 10))
	}

	if got := w.GetPreferredSize(); got != orvyn.NewSize(4, 12) {
		t.Errorf("preferred size = %v, want the explicit %v", got,
			orvyn.NewSize(4, 12))
	}
}

// An explicit minimum taller than the derived preferred height must not leave the
// widget asking the layout for less room than it insists on.
func TestPreferredHeightNeverDropsBelowAnExplicitMinimum(t *testing.T) {
	w := New("Description")
	w.SetContent([]string{"one short line"})
	w.SetMinSize(orvyn.NewSize(3, 10))

	if got := w.GetPreferredSize().Height; got < 10 {
		t.Errorf("preferred height = %d, below the explicit minimum of 10", got)
	}
}

// The styles used to stay zero-valued until a screen called OnFocus or OnBlur,
// so the widget rendered without its border and reported a frame of no height.
func TestFreshWidgetHasItsFrame(t *testing.T) {
	w := New("Description")

	if got := w.frameHeight(); got < 2 {
		t.Errorf("frame height = %d on a fresh widget; the border and title were "+
			"not accounted for", got)
	}
}
