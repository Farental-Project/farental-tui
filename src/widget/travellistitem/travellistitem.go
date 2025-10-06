package travellistitem

import (
	"farental/art"
	"farental/core/data/api"
	"farental/internal/style"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data *api.TravelResponse

	featuresList string

	style lipgloss.Style

	contentSize orvyn.Size
}

func Constructor(data *api.TravelResponse) list.ListItem {

	w := new(Widget)

	w.data = data

	w.UpdateData()

	w.OnBlur()

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 6

	w.BaseWidget.Resize(size)

	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height -= w.style.GetVerticalFrameSize()

	w.contentSize = size
}

func (w *Widget) UpdateData() {
	var b strings.Builder

	for _, f := range w.data.DestLocation.Features {
		if !f.IsAction {
			continue
		}

		if b.Len() > 0 {
			b.WriteString(fmt.Sprintf(" %c ", art.CharBullet))
		}

		b.WriteString(f.Name)
	}

	w.featuresList = b.String()
}

func (w *Widget) Render() string {
	t := orvyn.GetTheme()

	tui := w.style.Width(w.contentSize.Width).Render(
		fmt.Sprintf("%s\n%s\n%s\n%s",
			t.Style(theme.HighlightTextStyleID).
				Render(fmt.Sprintf("%s", w.data.DestLocation.Name)),
			style.LocationBiomeStyle(w.data.DestLocation.Biome.Code).
				Render(w.data.DestLocation.Biome.Name),
			t.Style(theme.NeutralDimTextStyleID).
				Render(w.data.DestLocation.Type.Name),
			t.Style(theme.DimTextStyleID).Render(w.featuresList),
		),
	)

	return tui
}

func (w *Widget) OnFocus() {
	w.style = orvyn.GetTheme().Style(theme.FocusedWidgetStyleID)
}

func (w *Widget) OnBlur() {
	w.style = orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)
}

func (w *Widget) OnEnterInput() {}

func (w *Widget) OnExitInput() {}

func (w *Widget) FilterValue() string {
	var b strings.Builder

	b.WriteString(w.data.DestLocation.Name)
	b.WriteString("-")
	b.WriteString(w.data.DestLocation.Continent.Name)
	b.WriteString("-")
	b.WriteString(w.data.DestLocation.Biome.Name)
	b.WriteString("-")
	b.WriteString(w.data.DestLocation.Type.Name)
	b.WriteString("-")
	b.WriteString(w.featuresList)

	return b.String()
}
