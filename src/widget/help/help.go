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

	w.BaseWidget = *orvyn.NewBaseWidget(w.Render)

	return w
}

func (w *Widget) Render() string {
	return bubblehelp.View(w.GetSize().Width)
}
