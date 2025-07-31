package travel

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

type Item struct {
	api.TravelResponse
}

func NewItem(travel *api.TravelResponse) Item {
	item := Item{
		TravelResponse: *travel,
	}

	return item
}

func (i Item) FilterValue() string {
	var b strings.Builder

	b.WriteString(i.ToLocation.Name)
	b.WriteString("\n")
	b.WriteString(i.ToLocation.Continent.Name)
	b.WriteString("\n")
	b.WriteString(i.ToLocation.Biome.Name)
	b.WriteString("\n")
	b.WriteString(i.ToLocation.Type.Name)

	return b.String()
}

type ItemDelegate struct{}

func (c ItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(Item)

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
			style.HighlightStyle.Render(fmt.Sprintf("%s", i.ToLocation.Name)),
			style.LocationBiomeStyle(i.ToLocation.Biome.Name).
				Render(i.ToLocation.Biome.Name),
			style.DimTextStyle.Render(i.ToLocation.Type.Name),
		),
	)

	fmt.Fprint(w, str)
}

func (c ItemDelegate) Height() int {
	return 3
}

func (c ItemDelegate) Spacing() int {
	return 0
}

func (c ItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
