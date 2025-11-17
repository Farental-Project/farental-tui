package travellistitem

import (
	"farental/art"
	"farental/core/data/api"
	"farental/internal/helper"
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

	data api.TravelResponse

	featuresList string
}

func Constructor(data api.TravelResponse) list.ListItem[api.TravelResponse] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.UpdateData(data)

	w.OnBlur()

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 5 + lipgloss.Height(w.featuresList)

	w.BaseWidget.Resize(size)
}

func (w *Widget) UpdateData(data api.TravelResponse) {
	var b strings.Builder

	t := orvyn.GetTheme()

	w.data = data

	for _, f := range w.data.DestLocation.Features {
		if b.Len() > 0 {
			b.WriteString(fmt.Sprintf(" %c ", art.CharBullet))
		}

		b.WriteString(f.Name)
	}

	w.featuresList = t.Style(theme.DimTextStyleID).Render(b.String())
}

func (w *Widget) GetData() api.TravelResponse {
	return w.data
}

func (w *Widget) Render() string {
	t := orvyn.GetTheme()
	contentSize := w.GetContentSize()

	width1, width2 := orvyn.DivideSizeFull(contentSize.Width)

	left := t.Style(theme.HighlightTextStyleID).
		Width(width1).AlignHorizontal(lipgloss.Left).
		Render(w.data.DestLocation.Name)
	right := t.Style(theme.NormalTextStyleID).
		Width(width2).AlignHorizontal(lipgloss.Right).
		Render(helper.HoursDecFormat(w.data.Duration))

	tui := w.GetStyle().Width(contentSize.Width).Render(
		fmt.Sprintf("%s\n%s\n%s\n%s",
			lipgloss.JoinHorizontal(lipgloss.Center, left, right),
			style.LocationBiomeStyle(w.data.DestLocation.Biome.Code).
				Render(w.data.DestLocation.Biome.Name),
			t.Style(theme.NeutralDimTextStyleID).
				Render(w.data.DestLocation.Type.Name),
			t.Style(theme.DimTextStyleID).Render(w.featuresList),
		),
	)

	return tui
}

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
