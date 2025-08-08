package equipmentsummary

import (
	"farental/internal/orvyn"
)

type Widget struct {
	orvyn.BaseWidget
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	return w
}
