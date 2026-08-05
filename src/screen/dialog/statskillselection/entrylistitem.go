package statskillselection

import (
	"fmt"

	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type listItem struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data Entry
}

func Constructor(data Entry) widgetlist.ListItem[Entry] {
	w := new(listItem)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.UpdateData(data)

	w.OnBlur()

	return w
}

func (w *listItem) Resize(size orvyn.Size) {
	size.Height = 3

	w.BaseWidget.Resize(size)
}

func (w *listItem) UpdateData(data Entry) {
	w.data = data
}

func (w *listItem) GetData() Entry {
	return w.data
}

func (w *listItem) FilterValue() string {
	return w.data.Label
}

func (w *listItem) Render() string {
	contentSize := w.GetContentSize()

	label := w.data.Label

	switch w.data.Kind {
	case EntryStat:
		label = fmt.Sprintf("%s %s", label,
			orvyn.GetTheme().Style(theme.DimTextStyleID).Render("("+lokyn.L("stat")+")"))
	case EntrySkill:
		label = fmt.Sprintf("%s %s", label,
			orvyn.GetTheme().Style(theme.DimTextStyleID).Render("("+lokyn.L("skill")+")"))
	}

	return w.GetStyle().Width(contentSize.Width).
		Height(contentSize.Height).
		Render(label)
}
