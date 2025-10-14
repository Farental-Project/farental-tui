package characterbasiclistitem

import (
	"farental/core/data/api"
	"farental/internal/style"
	"fmt"

	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data api.CharacterBasicResponse
}

func Constructor(data api.CharacterBasicResponse) list.ListItem[api.CharacterBasicResponse] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.data = data

	w.OnBlur()

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 5

	w.BaseWidget.Resize(size)
}

func (w *Widget) UpdateData(data api.CharacterBasicResponse) {
	w.data = data
}

func (w *Widget) GetData() api.CharacterBasicResponse {
	return w.data
}

func (w *Widget) Render() string {
	t := orvyn.GetTheme()
	contentSize := w.GetContentSize()

	str := w.GetStyle().Width(contentSize.Width).
		Height(contentSize.Height).Render(
		fmt.Sprintf("%s\n%s\n%s",
			t.Style(theme.HighlightTextStyleID).Render(fmt.Sprintf("%s %s", w.data.FirstName, w.data.LastName)),
			style.RaceStyle(w.data.RaceName).Render(w.data.RaceName),
			t.Style(theme.DimTextStyleID).Render(w.data.LocationName),
		),
	)

	return str
}

func (w *Widget) FilterValue() string {
	return ""
}
