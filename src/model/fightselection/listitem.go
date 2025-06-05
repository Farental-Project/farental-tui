package fightselection

import (
	"farental/core/data/api"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"strconv"
	"strings"
)

type ListItem struct {
	FightCompo api.FightCompositionResponse
	Paginator  paginator.Model
	TotalPower int
}

func NewListItem(f api.FightCompositionResponse) ListItem {
	li := ListItem{}

	li.FightCompo = f

	li.Paginator = paginator.New()
	li.Paginator.Type = paginator.Dots
	li.Paginator.PerPage = 4
	li.Paginator.ActiveDot = style.TitleStyle.Render("•")
	li.Paginator.InactiveDot = style.DimTextStyle.Render("•")
	li.Paginator.SetTotalPages(len(f.Actors))
	li.Paginator.KeyMap.NextPage = keybind.NextPage
	li.Paginator.KeyMap.PrevPage = keybind.PrevPage

	for _, a := range f.Actors {
		li.TotalPower += a.Power
	}

	return li
}

func (i ListItem) FilterValue() string {
	var b strings.Builder

	for _, a := range i.FightCompo.Actors {
		b.WriteString(a.Name)
	}

	return b.String()
}

type ListItemDelegate struct{}

func (l ListItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(ListItem)

	if !ok {
		return
	}

	var s lipgloss.Style
	var left strings.Builder
	var right strings.Builder
	var width int

	width = m.Width() - 2

	if index == m.Index() {
		s = style.FocusedStyle
	} else {
		s = style.BlurredStyle
	}

	perPage := i.Paginator.PerPage
	count := 0
	start, end := i.Paginator.GetSliceBounds(len(i.FightCompo.Actors))

	for i, pitem := range i.FightCompo.Actors[start:end] {
		if i > 0 {
			left.WriteString("\n")
		}

		left.WriteString(fmt.Sprintf("%s (%d)", pitem.Name, pitem.Power))

		count++
	}

	if count < perPage {
		left.WriteString(strings.Repeat("\n", perPage-count))
	}

	if len(i.FightCompo.Actors) > perPage {
		left.WriteString("\n")
		left.WriteString(style.HighlightStyle.Render("< "))
		left.WriteString(i.Paginator.View())
		left.WriteString(style.HighlightStyle.Render(" >"))
	} else {
		left.WriteString("\n")
	}

	right.WriteString(style.HighlightStyle.Render(strconv.Itoa(i.TotalPower)))
	right.WriteString("\n\n\n\n")
	right.WriteString(style.BoldTextStyle.Render(
		helper.HoursDecFormat(i.FightCompo.Duration.Duration)))

	tui := s.Width(m.Width() - 2).Height(l.Height()).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			style.TextStyle.Width(width/2).
				AlignHorizontal(lipgloss.Left).
				Render(left.String()),
			style.TextStyle.Width(width/2).
				AlignHorizontal(lipgloss.Right).
				Render(right.String())))

	fmt.Fprint(w, tui)
}

func (l ListItemDelegate) Height() int {
	return 5
}

func (l ListItemDelegate) Spacing() int {
	return 0
}

func (l ListItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	selectedIndex := m.Index()
	selectedItem, ok := m.SelectedItem().(ListItem)

	if !ok {
		return nil
	}

	selectedItem.Paginator, _ = selectedItem.Paginator.Update(msg)

	updateItem(m, selectedIndex, selectedItem)

	return nil
}

func updateItem(m *list.Model, index int, item ListItem) {
	m.SetItem(index, item)
}
