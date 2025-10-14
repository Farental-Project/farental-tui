package mailattachmentlistitem

import (
	"farental/core/data/api"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data api.StackResponse
}

func Constructor(data api.StackResponse) list.ListItem[api.StackResponse] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.data = data

	w.OnBlur()

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 3

	w.BaseWidget.Resize(size)
}

func (w *Widget) UpdateData(data api.StackResponse) {
	w.data = data
}

func (w *Widget) GetData() api.StackResponse {
	return w.data
}

func (w *Widget) Render() string {
	var s lipgloss.Style
	var b strings.Builder
	var width int

	width = w.GetContentSize().Width // borders

	s = w.GetStyle()

	b.WriteString(fmt.Sprintf("%dx %s", w.data.Count, w.data.Item.Name))

	tui := s.Width(width).Render(b.String())

	return tui
}

func (w *Widget) FilterValue() string {
	return ""
}
