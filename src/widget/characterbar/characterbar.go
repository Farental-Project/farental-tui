package characterbar

import (
	"farental/internal/orvyn"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Widget struct {
	orvyn.BaseWidget

	progress.Model

	TitleStyle lipgloss.Style

	title string

	MaxValue     int
	CurrentValue int
}

func New(title, color string) *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.Model = progress.New(progress.WithSolidFill(color))
	w.Model.ShowPercentage = false

	w.TitleStyle = style.NormalStyle.
		AlignHorizontal(lipgloss.Center)
	w.title = title

	return w
}

func (w *Widget) Render() string {
	var b strings.Builder

	size := w.GetSize()

	percent := float64((100 * w.CurrentValue / w.MaxValue) / 100)

	if len(w.title) > 0 {
		b.WriteString(w.TitleStyle.Render(
			fmt.Sprintf("%s (%d/%d)",
				w.title, w.CurrentValue, w.MaxValue)))
	} else {
		b.WriteString(w.TitleStyle.Render(fmt.Sprintf("(%d/%d)",
			w.CurrentValue, w.MaxValue)))
	}
	b.WriteString(strings.Repeat(
		fmt.Sprintf("\n%s", w.Model.ViewAs(percent)), size.Height-1),
	)

	return b.String()
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	w.Model.Width = size.Width

	w.TitleStyle = w.TitleStyle.Width(size.Width)
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(6, 4)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(6, 4)
}
