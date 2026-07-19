package characterbasiclistitem

import (
	"farental/core/data/api"
	"farental/internal/style"
	"fmt"
	"strings"

	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Data struct {
	api.CharacterBasicResponse
	ShowLocation bool
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data Data
}

func Constructor(data Data) widgetlist.ListItem[Data] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.data = data

	w.OnBlur()

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	if w.data.ShowLocation {
		size.Height = 5
	} else {
		size.Height = 4
	}

	w.BaseWidget.Resize(size)
}

func (w *Widget) UpdateData(data Data) {
	w.data = data
}

func (w *Widget) GetData() Data {
	return w.data
}

func (w *Widget) Render() string {
	var b strings.Builder

	t := orvyn.GetTheme()
	contentSize := w.GetContentSize()

	b.WriteString(t.Style(theme.HighlightTextStyleID).Render(
		fmt.Sprintf("%s %s", w.data.FirstName, w.data.LastName)))
	b.WriteString("\n")
	fmt.Fprintf(&b, "%s - %s",
		style.RaceStyle(w.data.RaceName).Render(w.data.RaceName),
		w.data.Gender)

	if w.data.ShowLocation {
		b.WriteString("\n")
		b.WriteString(t.Style(theme.DimTextStyleID).Render(w.data.LocationName))
	}

	str := b.String()

	return w.GetStyle().
		Width(contentSize.Width).
		Height(contentSize.Height).
		Render(str)
}

func (w *Widget) FilterValue() string {
	var b strings.Builder

	b.WriteString(w.data.FirstName)
	b.WriteString(" ")
	b.WriteString(w.data.LastName)
	b.WriteString(" ")
	b.WriteString(w.data.Gender)
	b.WriteString(" ")
	b.WriteString(w.data.RaceName)

	return b.String()
}
