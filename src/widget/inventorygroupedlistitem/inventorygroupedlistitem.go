package inventorygroupedlistitem

import (
	"farental/core/data/api"
	"farental/internal/helper"
	"farental/internal/keybind"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Data struct {
	api.ItemResponse
	StackCount int
	Count      int
	Amount     int
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data Data

	style lipgloss.Style

	contentSize orvyn.Size
}

func Constructor(data Data) list.ListItem[Data] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.data = data

	w.OnBlur()

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	switch msgType := msg.(type) {
	case tea.KeyMsg:

		switch {
		case key.Matches(msgType, keybind.Left):
			w.data.Amount--

			w.data.Amount = max(0, w.data.Amount)

		case key.Matches(msgType, keybind.ShiftLeft):
			w.data.Amount -= helper.Prev10Inc(w.data.Amount)

			w.data.Amount = max(0, w.data.Amount)

		case key.Matches(msgType, keybind.Right):
			w.data.Amount++

			w.data.Amount = min(w.data.Amount, w.data.Count)

		case key.Matches(msgType, keybind.ShiftRight):
			w.data.Amount += helper.Next10Inc(w.data.Amount)

			w.data.Amount = min(w.data.Amount, w.data.Count)

		}
	}

	return nil
}

func (w *Widget) UpdateData(data Data) {
	w.data = data
}

func (w *Widget) GetData() Data {
	return w.data
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
	hs := t.Style(theme.HighlightTextStyleID)

	left.WriteString(w.data.Name)
	left.WriteString("\n")
	left.WriteString(t.Style(theme.DimTextStyleID).Render(
		fmt.Sprintf(lokyn.L("Stack count : %d"), w.data.StackCount)),
	)

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
