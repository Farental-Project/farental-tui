package inventorylistitem

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

	data api.StackResponse

	contentSize orvyn.Size
}

func Constructor(data api.StackResponse) list.ListItem[api.StackResponse] {
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

func (w *Widget) UpdateData(data api.StackResponse) {
	w.data = data
}

func (w *Widget) GetData() api.StackResponse {
	return w.data
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

	left.WriteString(w.data.Item.Name)

	right.WriteString(t.Style(theme.DimTextStyleID).Render(
		fmt.Sprintf("%d / %d", w.data.Count, w.data.Item.MaxStackCount)))

	width1, width2 := orvyn.DivideSizeFull(width)

	tui := s.Width(width).Height(w.contentSize.Height).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			ns.Width(width1).
				AlignHorizontal(lipgloss.Left).
				Render(left.String()),
			ns.Width(width2).
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

	b.WriteString(w.data.Item.Name)
	b.WriteString(" ")
	b.WriteString(w.data.Item.Description)

	if w.data.Item.EquipmentSlot != nil {
		b.WriteString(w.data.Item.EquipmentSlot.Name)
	}

	return b.String()
}
