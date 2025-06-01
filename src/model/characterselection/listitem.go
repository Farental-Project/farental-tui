package characterselection

import (
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
)

type ListItem struct {
	CharacterID       uint
	CharacterName     string
	CharacterRace     string
	CharacterLocation string
}

func (i ListItem) FilterValue() string { return "" }

type ListItemDelegate struct {
}

func (l ListItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(ListItem)

	if !ok {
		return
	}

	var s lipgloss.Style

	if index == m.Index() {
		s = style.FocusedStyle
	} else {
		s = style.BlurredStyle
	}

	str := s.Width(m.Width()).Render(fmt.Sprintf(
		"%s\n%s\n%s",
		style.HighlightStyle.Render(i.CharacterName),
		style.RaceStyle(i.CharacterRace).Render(i.CharacterRace),
		style.DimTextStyle.Render(i.CharacterLocation)))

	fmt.Fprint(w, str)
}

func (l ListItemDelegate) Height() int {
	return 3
}

func (l ListItemDelegate) Spacing() int {
	return 0
}

func (l ListItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
