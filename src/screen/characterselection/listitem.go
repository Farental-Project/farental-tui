package characterselection

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

type CharacterItem struct {
	api.CharacterBasicResponse
}

func NewCharacterItem(character *api.CharacterBasicResponse) CharacterItem {
	characterItem := CharacterItem{
		CharacterBasicResponse: *character,
	}

	return characterItem
}

func (i CharacterItem) FilterValue() string {
	var b strings.Builder

	b.WriteString(i.FirstName)
	b.WriteString(" ")
	b.WriteString(i.LastName)
	b.WriteString(" ")
	b.WriteString(i.RaceName)
	b.WriteString(" ")
	b.WriteString(i.LocationName)

	return b.String()
}

type CharacterItemDelegate struct{}

func (c CharacterItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(CharacterItem)

	if !ok {
		return
	}

	var s lipgloss.Style

	if index == m.Index() {
		s = style.FocusedStyle
	} else {
		s = style.BlurredStyle
	}

	str := s.Width(m.Width()).Render(
		fmt.Sprintf("%s\n%s\n%s",
			style.HighlightStyle.Render(fmt.Sprintf("%s %s", i.FirstName, i.LastName)),
			style.RaceStyle(i.RaceName).Render(i.RaceName),
			style.DimTextStyle.Render(i.LocationName),
		),
	)

	fmt.Fprint(w, str)
}

func (c CharacterItemDelegate) Height() int {
	return 3
}

func (c CharacterItemDelegate) Spacing() int {
	return 0
}

func (c CharacterItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
