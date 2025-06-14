package inventory

import (
	"farental/core/data/api"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"strings"
)

type ListItem struct {
	Stack api.StackResponse
}

func (i ListItem) FilterValue() string {
	var b strings.Builder

	b.WriteString(i.Stack.Item.Name)
	b.WriteString(i.Stack.Item.Description)
	
	if i.Stack.Item.EquipmentSlot != nil {
		b.WriteString(i.Stack.Item.EquipmentSlot.Name)
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

	left.WriteString(i.Stack.Item.Name)
	left.WriteString("\n")
	left.WriteString(style.DimTextStyle.Render(i.Stack.Item.Description))

	right.WriteString(style.DimTextStyle.Render(
		fmt.Sprintf("%d / %d", i.Stack.Count, i.Stack.Item.MaxStackCount)))

	if i.Stack.Item.EquipmentSlot != nil {
		right.WriteString(i.Stack.Item.EquipmentSlot.Name)
		right.WriteString("\n")
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
	selectedIndex := m.Index()
	selectedItem, ok := m.SelectedItem().(ListItem)

	if !ok {
		return nil
	}

	// switch msg := msg.(type) {
	// case tea.KeyMsg:
	// 	return nil
	// }

	updateItem(m, selectedIndex, selectedItem)
	return nil
}

func updateItem(m *list.Model, index int, item ListItem) {
	m.SetItem(index, item)
}
