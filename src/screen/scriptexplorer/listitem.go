package scriptexplorer

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
	api.ScriptBasicResponse
}

func NewListItem(script api.ScriptBasicResponse) ListItem {
	return ListItem{script}
}

func (i ListItem) FilterValue() string {
	var b strings.Builder

	b.WriteString(i.Name)
	b.WriteString(" ")
	b.WriteString(i.Description)

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

	left.WriteString(style.TitleStyle.Render(i.Name))
	left.WriteString("\n")
	left.WriteString(style.DimTextStyle.Render(i.Description))

	tui := s.Width(m.Width() - 2).Height(l.Height()).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			style.TextStyle.Width(width-2).
				AlignHorizontal(lipgloss.Left).
				Render(left.String()),
			style.TextStyle.Width(1).
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
	return nil
}
