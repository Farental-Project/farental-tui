package fightlistitem

import (
	"farental/core/data/api"
	"farental/internal/helper"
	"farental/internal/keybind"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Data struct {
	api.FightCompositionResponse
	TotalPower int
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data *Data

	paginator paginator.Model

	style lipgloss.Style

	contentSize orvyn.Size
}

func Constructor(data *Data) list.ListItem {
	w := new(Widget)

	w.data = data

	w.paginator = paginator.New()
	w.paginator.Type = paginator.Dots
	w.paginator.PerPage = 4
	widget.UpdatePaginatorTheme(&w.paginator)
	w.paginator.SetTotalPages(len(data.Actors))
	w.paginator.KeyMap.NextPage = keybind.Right
	w.paginator.KeyMap.PrevPage = keybind.Left

	w.UpdateData()

	w.OnBlur()

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	w.paginator, _ = w.paginator.Update(msg)

	return nil
}

func (w *Widget) UpdateData() {
	for _, a := range w.data.Actors {
		w.data.TotalPower += a.Power
	}
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 7

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
	hs := t.Style(theme.HighlightTextStyleID)
	ns := lipgloss.NewStyle()

	perPage := w.paginator.PerPage
	count := 0
	start, end := w.paginator.GetSliceBounds(len(w.data.Actors))

	for i, a := range w.data.Actors[start:end] {
		if i > 0 {
			left.WriteString("\n")
		}

		left.WriteString(fmt.Sprintf("%s (%d)", a.Name, a.Power))

		count++
	}

	if count < perPage {
		left.WriteString(strings.Repeat("\n", perPage-count))
	}

	if len(w.data.Actors) > perPage {
		left.WriteString("\n")
		left.WriteString(hs.Render("< "))
		left.WriteString(w.paginator.View())
		left.WriteString(hs.Render(" >"))
	} else {
		left.WriteString("\n")
	}

	right.WriteString(hs.Render(strconv.Itoa(w.data.TotalPower)))
	right.WriteString("\n\n\n\n")
	right.WriteString(t.Style(theme.NeutralTextStyleID).Bold(true).Render(
		helper.HoursDecFormat(w.data.Duration.Duration)))

	tui := s.Width(width).Height(w.contentSize.Height).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
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

	for _, a := range w.data.Actors {
		b.WriteString(a.Name)
	}

	return b.String()
}
