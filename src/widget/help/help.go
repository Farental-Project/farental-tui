package help

import (
	"farental/internal/orvyn"
	"github.com/halsten-dev/bubblehelp"
)

type Widget struct {
	orvyn.BaseWidget
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = *orvyn.NewBaseWidget()

	return w
}

func (w *Widget) Render(size orvyn.Size) string {
	return bubblehelp.View(size.Width)
}
