package scriptexplorerlistitem

import (
	"farental/core/data/api"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
	"strings"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data *api.ScriptBasicResponse

	style lipgloss.Style

	contentSize orvyn.Size
}

func Constructor(data *api.ScriptBasicResponse) list.IListItem {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.data = data

	w.OnBlur()

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 4

	w.BaseWidget.Resize(size)

	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height -= w.style.GetVerticalFrameSize()

	w.contentSize = size
}

func (w *Widget) Render() string {
	var s lipgloss.Style
	var left strings.Builder
	var right strings.Builder
	var width int

	width = w.contentSize.Width

	s = w.style
	t := orvyn.GetTheme()
	ns := lipgloss.NewStyle()

	left.WriteString(t.Style(theme.TitleStyleID).Render(w.data.Name))
	left.WriteString("\n")
	left.WriteString(t.Style(theme.DimTextStyleID).Render(w.data.Description))

	tui := s.Width(width).Height(w.contentSize.Height).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			ns.Width(width-2).
				AlignHorizontal(lipgloss.Left).
				Render(left.String()),
			ns.Width(1).
				AlignHorizontal(lipgloss.Right).
				Render(right.String())))

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

	b.WriteString(w.data.Name)
	b.WriteString(" ")
	b.WriteString(w.data.Description)

	return b.String()
}
