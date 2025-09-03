package mailattachmentselectlistitem

import (
	"farental/core/data/api"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
	"strconv"
	"strings"
)

type Data struct {
	api.ItemResponse
	Count  int
	Amount int
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data *Data

	style lipgloss.Style

	contentSize orvyn.Size
}

func Constructor(data *Data) list.IListItem {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.data = data

	w.OnBlur()

	return w
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
	hs := t.Style(theme.HighlightTextStyleID)

	left.WriteString(w.data.Name)

	right.WriteString(t.Style(theme.DimTextStyleID).Render(
		fmt.Sprintf("%d", w.data.Count),
	))
	right.WriteString("\n")

	// Amount control
	right.WriteString(hs.Render("< "))
	right.WriteString(t.Style(theme.NeutralTextStyleID).
		Render(strconv.Itoa(w.data.Amount)))
	right.WriteString(hs.Render(" >"))

	tui := s.Render(lipgloss.JoinHorizontal(lipgloss.Top,
		ns.Width(width/2).
			AlignHorizontal(lipgloss.Left).
			Render(left.String()),
		ns.Width(width/2).
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

	return b.String()
}
