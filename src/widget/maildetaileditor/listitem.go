package maildetaileditor

import (
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"strings"
)

type ListItem struct {
	StackID  uint
	ItemID   uint
	ItemName string
	Amount   int
}

func (i ListItem) FilterValue() string {
	return ""
}

type ListItemDelegate struct {
}

func (l ListItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(ListItem)

	if !ok {
		return
	}

	var s lipgloss.Style
	var b strings.Builder
	var width int

	width = m.Width() - 2 // borders

	if index == m.Index() {
		s = style.FocusedStyle
	} else {
		s = style.BlurredStyle
	}

	b.WriteString(fmt.Sprintf("%dx %s", i.Amount, i.ItemName))

	tui := s.Width(width).Render(b.String())

	fmt.Fprint(w, tui)

}

func (l ListItemDelegate) Height() int {
	return 1
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

	// TODO : Manage delete
	// switch msg := msg.(type) {
	// case tea.KeyMsg:
	// 	switch {
	//
	// 	}
	// }

	updateItem(m, selectedIndex, selectedItem)
	return nil
}

func updateItem(m *list.Model, index int, item ListItem) {
	m.SetItem(index, item)
}
