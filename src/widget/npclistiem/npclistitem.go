package npclistitem

import (
	"farental/core/data/api"
	"farental/internal/style"
	"fmt"

	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data api.NpcResponse
}

func Constructor(data api.NpcResponse) widgetlist.ListItem[api.NpcResponse] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.data = data

	w.OnBlur()

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 4

	w.BaseWidget.Resize(size)
}

func (w *Widget) UpdateData(data api.NpcResponse) {
	w.data = data
}

func (w *Widget) GetData() api.NpcResponse {
	return w.data
}

func (w *Widget) Render() string {
	t := orvyn.GetTheme()
	contentSize := w.GetContentSize()

	name := ""

	if w.data.Name != "" {
		name = w.data.Name
	} else {
		name = fmt.Sprintf("%s %s", w.data.FirstName, w.data.LastName)
	}

	str := w.GetStyle().Width(contentSize.Width).
		Height(contentSize.Height).Render(
		fmt.Sprintf("%s\n%s",
			t.Style(theme.HighlightTextStyleID).Render(name),
			style.RaceStyle(w.data.RaceName).Render(w.data.RaceName),
		),
	)

	return str
}

func (w *Widget) FilterValue() string {
	return ""
}
