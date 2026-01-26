package fighthistorylistitem

import (
	"farental/core/data/api"
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
	"github.com/halsten-dev/orvyn/widget/widgetlist"
	"github.com/spf13/viper"
)

type Data struct {
	api.FightResponse
	TotalPower int
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data Data

	paginator paginator.Model
}

func Constructor(data Data) widgetlist.ListItem[Data] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.UpdateData(data)

	w.paginator = paginator.New()
	w.paginator.Type = paginator.Dots
	w.paginator.PerPage = 4
	widget.UpdatePaginatorTheme(&w.paginator)
	w.paginator.SetTotalPages(len(data.Composition.Actors))
	w.paginator.KeyMap.NextPage = keybind.Right
	w.paginator.KeyMap.PrevPage = keybind.Left

	w.OnBlur()

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	w.paginator, _ = w.paginator.Update(msg)

	return nil
}

func (w *Widget) UpdateData(data Data) {
	w.data = data

	for _, a := range w.data.Composition.Actors {
		w.data.TotalPower += a.Power
	}
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
	start, end := w.paginator.GetSliceBounds(len(w.data.Composition.Actors))

	for i, a := range w.data.Composition.Actors[start:end] {
		if i > 0 {
			left.WriteString("\n")
		}

		fmt.Fprintf(&left, "%s (%d)", a.Name, a.Power)

		count++
	}

	if count < perPage {
		left.WriteString(strings.Repeat("\n", perPage-count))
	}

	if len(w.data.Composition.Actors) > perPage {
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
		w.data.ResolvedTimestamp.Format(viper.GetString("datetimeformat"))),
	)

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

	for _, a := range w.data.Composition.Actors {
		b.WriteString(a.Name)
	}

	return b.String()
}
