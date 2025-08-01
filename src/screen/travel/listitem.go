package travel

import (
	"farental/art"
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

	featuresList string
}

func NewItem(travel *api.TravelResponse) Item {
	var b strings.Builder

	item := Item{
		TravelResponse: *travel,
	}

	for _, f := range item.DestLocation.Features {
		if !f.IsAction {
			continue
		}

		if b.Len() > 0 {
			b.WriteString(fmt.Sprintf(" %c ", art.CharBullet))
		}

		b.WriteString(f.Name)
	}

	item.featuresList = b.String()

	return item
}

func (i Item) FilterValue() string {
	var b strings.Builder

	b.WriteString(i.DestLocation.Name)
	b.WriteString("\n")
	b.WriteString(i.DestLocation.Continent.Name)
	b.WriteString("\n")
	b.WriteString(i.DestLocation.Biome.Name)
	b.WriteString("\n")
	b.WriteString(i.DestLocation.Type.Name)

	for i, f := range i.DestLocation.Features {
		if i > 0 {
			b.WriteString("\n")
		}

		b.WriteString(f.Name)
	}

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
		fmt.Sprintf("%s\n%s\n%s\n%s",
			style.HighlightStyle.Render(fmt.Sprintf("%s", i.DestLocation.Name)),
			style.LocationBiomeStyle(i.DestLocation.Biome.Name).
				Render(i.DestLocation.Biome.Name),
			style.NeutralDimTextStyle.Render(i.DestLocation.Type.Name),
			style.DimTextStyle.Render(i.featuresList),
		),
	)

	fmt.Fprint(w, str)
}

func (c ItemDelegate) Height() int {
	return 4
}

func (c ItemDelegate) Spacing() int {
	return 0
}

func (c ItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
