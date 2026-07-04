package inventorylistitem

import (
	"farental/core/data/api"
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data api.StackResponse

	stackCount int
}

func Constructor(data api.StackResponse) widgetlist.ListItem[api.StackResponse] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.UpdateData(data)

	w.OnBlur()

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 3

	w.BaseWidget.Resize(size)
}

func (w *Widget) UpdateData(data api.StackResponse) {
	w.data = data

	w.stackCount = int(math.Ceil(float64(w.data.Count) / float64(w.data.Item.MaxStackCount)))
}

func (w *Widget) GetData() api.StackResponse {
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

	left.WriteString(w.data.Item.Name)

	if w.data.Count > 0 {
		right.WriteString(t.Style(theme.DimTextStyleID).Render(
			fmt.Sprintf("x%d | %d %s", w.data.Count, w.stackCount, lokyn.L("Stacks"))))
	}

	width1, width2 := orvyn.DivideSizeFull(width)

	tui := s.Width(width).Height(contentSize.Height).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			ns.Width(width1).
				AlignHorizontal(lipgloss.Left).
				Render(left.String()),
			ns.Width(width2).
				AlignHorizontal(lipgloss.Right).
				Render(right.String())))

	return tui
}

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
