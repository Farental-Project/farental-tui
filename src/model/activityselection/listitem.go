package activityselection

import (
	"farental/core/data/api"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"strings"
)

type ListItem struct {
	Activity      api.ActivityResponse
	DurationIndex int
}

func (i ListItem) FilterValue() string {
	var b strings.Builder

	b.WriteString(i.Activity.Name)
	b.WriteString(i.Activity.Description)
	b.WriteString(i.Activity.Skill.Name)

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

	left.WriteString(i.Activity.Name)
	left.WriteString("\n")
	left.WriteString(style.DimTextStyle.Render(i.Activity.Description))

	right.WriteString(style.DimTextStyle.Render(i.Activity.Skill.Name))
	right.WriteString("\n\n\n")

	if len(i.Activity.Duration.Durations) > 0 {
		right.WriteString(style.HighlightStyle.Render("< "))
		right.WriteString(style.BoldTextStyle.Render(helper.HoursDecFormat(i.Activity.Duration.Durations[i.DurationIndex].Duration)))
		right.WriteString(style.HighlightStyle.Render(" >"))
	} else {
		right.WriteString(style.BoldTextStyle.Render(helper.HoursDecFormat(i.Activity.Duration.Durations[0].Duration)))
	}

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
	return 4
}

func (l ListItemDelegate) Spacing() int {
	return 0
}

func (l ListItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	selectedIndex := m.GlobalIndex()
	selectedItem, ok := m.SelectedItem().(ListItem)

	if !ok {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch {
		case key.Matches(msg, keybind.Left):
			selectedItem.DurationIndex--

			if selectedItem.DurationIndex < 0 {
				selectedItem.DurationIndex = 0
			}

			updateItem(m, selectedIndex, selectedItem)
		case key.Matches(msg, keybind.Right):
			selectedItem.DurationIndex++

			length := len(selectedItem.Activity.Duration.Durations) - 1

			if selectedItem.DurationIndex > length {
				selectedItem.DurationIndex = length
			}

			updateItem(m, selectedIndex, selectedItem)
		}
	}
	return nil
}

func updateItem(m *list.Model, index int, item ListItem) {
	m.SetItem(index, item)
}
