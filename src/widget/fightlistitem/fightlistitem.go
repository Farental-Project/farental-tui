package fightlistitem

import (
	"farental/core/data/api"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/widget/numericalselector"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Data struct {
	api.FightCompositionResponse
	TotalPower int
	TotalTime  float64
	Amount     int
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	amountSelector *numericalselector.Widget
	paginator      paginator.Model

	data Data
}

func Constructor(data Data) widgetlist.ListItem[Data] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.amountSelector = numericalselector.New(1, 100, 1)
	w.paginator = paginator.New()
	w.paginator.Type = paginator.Dots
	w.paginator.PerPage = 4
	widget.UpdatePaginatorTheme(&w.paginator)
	w.paginator.SetTotalPages(len(data.Actors))
	w.paginator.KeyMap.NextPage = keybind.NextPage
	w.paginator.KeyMap.PrevPage = keybind.PrevPage

	w.UpdateData(data)

	w.OnBlur()

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	if w.amountSelector.IsActive() {
		w.amountSelector.Update(msg)
	}

	w.paginator, _ = w.paginator.Update(msg)

	w.data.Amount = w.amountSelector.GetValue()

	w.recalcTotalPower()
	w.recalcTotalTime()

	return nil
}

func (w *Widget) recalcTotalTime() {
	w.data.TotalTime = w.data.FightCompositionResponse.Duration.Duration * float64(w.data.Amount)
}

func (w *Widget) recalcTotalPower() {
	w.data.TotalPower = 0

	for _, a := range w.data.Actors {
		w.data.TotalPower += a.Power
	}

	w.data.TotalPower *= w.data.Amount
}

func (w *Widget) UpdateData(data Data) {
	w.data = data

	w.recalcTotalPower()

	w.amountSelector.SetActive(w.data.FightCompositionResponse.Simple)
}

func (w *Widget) GetData() Data {
	return w.data
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 7

	w.BaseWidget.Resize(size)
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
	hs := t.Style(theme.HighlightTextStyleID)
	ns := lipgloss.NewStyle()

	perPage := w.paginator.PerPage
	count := 0
	start, end := w.paginator.GetSliceBounds(len(w.data.Actors))

	for i, a := range w.data.Actors[start:end] {
		if i > 0 {
			left.WriteString("\n")
		}

		fmt.Fprintf(&left, "%s (%d)", a.Name, a.Power)

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
	right.WriteString("\n")
	right.WriteString(t.Style(theme.NeutralTextStyleID).Bold(true).Render(
		helper.HoursDecFormat(w.data.TotalTime)))
	right.WriteString("\n\n\n")

	if w.amountSelector.IsActive() {
		right.WriteString(w.amountSelector.Render())
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

	for _, a := range w.data.Actors {
		b.WriteString(a.Name)
	}

	return b.String()
}
