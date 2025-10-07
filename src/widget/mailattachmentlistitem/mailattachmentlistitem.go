package mailattachmentlistitem

import (
	"farental/core/data/api"
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

	style lipgloss.Style

	data *api.StackResponse

	contentSize orvyn.Size
}

func Constructor(data *api.StackResponse) list.ListItem {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.data = data

	w.OnBlur()

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 3

	w.BaseWidget.Resize(size)

	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height -= w.style.GetVerticalFrameSize()

	w.contentSize = size
}

func (w *Widget) UpdateData() {}

func (w *Widget) Render() string {
	var s lipgloss.Style
	var b strings.Builder
	var width int

	width = w.contentSize.Width // borders

	s = w.style

	b.WriteString(fmt.Sprintf("%dx %s", w.data.Count, w.data.Item.Name))

	tui := s.Width(width).Render(b.String())

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
	return ""
}
