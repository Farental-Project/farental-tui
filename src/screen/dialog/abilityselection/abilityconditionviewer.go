package abilityselection

import (
	"farental/internal/keybind"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget"
)

type AbilityConditionViewer struct {
	orvyn.BaseWidget

	paginator paginator.Model

	data []string
}

func NewAbilityConditionViewer(data []string) *AbilityConditionViewer {
	w := new(AbilityConditionViewer)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.data = data

	w.paginator = paginator.New()
	w.paginator.Type = paginator.Dots
	w.paginator.PerPage = 6
	widget.UpdatePaginatorTheme(&w.paginator)
	w.paginator.SetTotalPages(len(w.data))
	w.paginator.KeyMap.NextPage = keybind.Right
	w.paginator.KeyMap.PrevPage = keybind.Left

	w.BaseWidget.SetStyle(lipgloss.NewStyle())

	return w
}

func (a *AbilityConditionViewer) Update(msg tea.Msg) tea.Cmd {
	a.paginator, _ = a.paginator.Update(msg)

	return nil
}

func (a *AbilityConditionViewer) Resize(size orvyn.Size) {
	a.BaseWidget.Resize(size)
}

func (a *AbilityConditionViewer) Render() string {
	var list, left, right strings.Builder
	var b *strings.Builder
	var width int

	size := a.GetContentSize()
	t := orvyn.GetTheme()
	hs := t.Style(theme.HighlightTextStyleID)
	ds := t.Style(theme.DimTextStyleID)
	ns := lipgloss.NewStyle()

	width = size.Width

	if len(a.data) == 0 {
		return ds.Width(width).Render(lokyn.L("No conditions"))
	}

	width1, width2 := orvyn.DivideSizeFull(width)

	perPage := a.paginator.PerPage
	count := 0
	start, end := a.paginator.GetSliceBounds(len(a.data))

	for i, condition := range a.data[start:end] {
		if i%2 == 0 {
			b = &left
		} else {
			b = &right
		}

		if i > 1 {
			b.WriteString("\n")
		}

		b.WriteString(condition)

		count++
	}

	list.WriteString(
		lipgloss.JoinHorizontal(lipgloss.Top,
			ns.Width(width1).
				AlignHorizontal(lipgloss.Left).
				Render(left.String()),
			ns.Width(width2).
				AlignHorizontal(lipgloss.Left).
				Render(right.String())))

	if count < 4 {
		list.WriteString(strings.Repeat("\n", 4-count))
	}

	if len(a.data) > perPage {
		list.WriteString("\n")
		list.WriteString(hs.Render("< "))
		list.WriteString(a.paginator.View())
		list.WriteString(hs.Render(" >"))
	}

	return ds.Render(list.String())
}
