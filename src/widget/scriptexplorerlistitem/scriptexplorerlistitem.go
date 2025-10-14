package scriptexplorerlistitem

import (
	"farental/core/data/api"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data api.ScriptBasicResponse
}

func Constructor(data api.ScriptBasicResponse) list.ListItem[api.ScriptBasicResponse] {
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

func (w *Widget) UpdateData(data api.ScriptBasicResponse) {
	w.data = data
}

func (w *Widget) GetData() api.ScriptBasicResponse {
	return w.data
}

func (w *Widget) Render() string {
	var s lipgloss.Style
	var left strings.Builder
	var right strings.Builder
	var width int

	contentSize := w.GetContentSize()
	width = contentSize.Width

	s = w.GetStyle()
	t := orvyn.GetTheme()
	ns := lipgloss.NewStyle()

	left.WriteString(t.Style(theme.TitleStyleID).Render(w.data.Name))
	left.WriteString("\n")
	left.WriteString(t.Style(theme.DimTextStyleID).Render(w.data.Description))

	tui := s.Width(width).Height(contentSize.Height).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			ns.Width(width-2).
				AlignHorizontal(lipgloss.Left).
				Render(left.String()),
			ns.Width(1).
				AlignHorizontal(lipgloss.Right).
				Render(right.String())))

	return tui
}

func (w *Widget) FilterValue() string {
	var b strings.Builder

	b.WriteString(w.data.Name)
	b.WriteString(" ")
	b.WriteString(w.data.Description)

	return b.String()
}
