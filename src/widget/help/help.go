package help

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/orvyn"
)

type Widget struct {
	orvyn.BaseWidget
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	return w
}

func (w *Widget) Render() string {
	return bubblehelp.View(w.GetSize().Width)
}

// GetMinSize reports a fixed width of 1 so the help bar never drives the
// layout width. Deriving the width from bubblehelp.View(w.GetSize().Width)
// is self-referential: View renders into .Width(currentWidth), so its measured
// width equals the current width. A width-based layout (e.g. VBoxLayout) would
// then size the form from it, resize the help smaller, and read a smaller width
// on the next frame - shrinking the form one margin per render/keypress.
// Only the height depends on the (layout-assigned) render width.
func (w *Widget) GetMinSize() orvyn.Size {
	h := orvyn.GetRenderSize(lipgloss.NewStyle(), bubblehelp.View(w.GetSize().Width)).Height
	return orvyn.NewSize(1, h)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return w.GetMinSize()
}
