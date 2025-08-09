package mailattachmentselect

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
	"strconv"
	"strings"
)

type ListItem struct {
	Stack  api.StackResponse
	Amount int
}

func NewListItem(stack *api.StackResponse) ListItem {
	li := ListItem{}

	li.Stack = *stack
	li.Amount = 0

	return li
}

func (i ListItem) FilterValue() string {
	var b strings.Builder

	b.WriteString(i.Stack.Item.Name)
	b.WriteString(i.Stack.Item.Description)

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

	left.WriteString(i.Stack.Item.Name)

	right.WriteString(style.DimTextStyle.Render(
		fmt.Sprintf("%d / %d", i.Stack.Count, i.Stack.Item.MaxStackCount),
	))
	right.WriteString("\n")

	// Amount control
	right.WriteString(style.HighlightStyle.Render("< "))
	right.WriteString(style.BoldTextStyle.Render(strconv.Itoa(i.Amount)))
	right.WriteString(style.HighlightStyle.Render(" >"))

	tui := s.Render(lipgloss.JoinHorizontal(lipgloss.Top,
		style.TextStyle.Width(width/2).
			AlignHorizontal(lipgloss.Left).
			Render(left.String()),
		style.TextStyle.Width(width/2).
			AlignHorizontal(lipgloss.Right).
			Render(right.String())))

	fmt.Fprint(w, tui)
}

func (l ListItemDelegate) Height() int {
	return 2
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

	switch msgType := msg.(type) {
	case tea.KeyMsg:

		switch {
		case key.Matches(msgType, keybind.Left):
			selectedItem.Amount--

			selectedItem.Amount = max(0, selectedItem.Amount)

			return updateItem(m, selectedIndex, selectedItem)
		case key.Matches(msgType, keybind.ShiftLeft):
			selectedItem.Amount -= helper.Prev10Inc(selectedItem.Amount)

			selectedItem.Amount = max(0, selectedItem.Amount)

			return updateItem(m, selectedIndex, selectedItem)
		case key.Matches(msgType, keybind.Right):
			selectedItem.Amount++

			selectedItem.Amount = min(selectedItem.Amount, selectedItem.Stack.Count)

			return updateItem(m, selectedIndex, selectedItem)
		case key.Matches(msgType, keybind.ShiftRight):
			selectedItem.Amount += helper.Next10Inc(selectedItem.Amount)

			selectedItem.Amount = min(selectedItem.Amount, selectedItem.Stack.Count)

			return updateItem(m, selectedIndex, selectedItem)
		}
	}

	return updateItem(m, selectedIndex, selectedItem)
}

func updateItem(m *list.Model, index int, item ListItem) tea.Cmd {
	return m.SetItem(index, item)
}
