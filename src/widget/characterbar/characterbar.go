package characterbar

import (
	"farental/internal/orvyn"
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	"strings"
)

type Widget struct {
	orvyn.BaseWidget

	progress.Model

	MaxValue     int
	CurrentValue int
}

func New(color string) *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.Model = progress.New(progress.WithSolidFill(color))
	w.Model.ShowPercentage = false

	return w
}

func (w *Widget) Render() string {
	var b strings.Builder

	size := w.GetSize()

	percent := float64((100 * w.CurrentValue / w.MaxValue) / 100)

	b.WriteString(fmt.Sprintf("(%d/%d)",
		w.CurrentValue, w.MaxValue))
	b.WriteString(strings.Repeat(
		fmt.Sprintf("\n%s", w.Model.ViewAs(percent)), size.Height-1),
	)

	return b.String()
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	w.Model.Width = size.Width
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(6, 4)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(6, 4)
}
