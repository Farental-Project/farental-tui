package travelselection

import (
	"farental/core/data/api"
	"farental/internal/helper"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"strings"
)

type ListItem struct {
	Travel api.TravelResponse
}

func (i ListItem) FilterValue() string {
	return i.Travel.DestLocation.Name
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

	left.WriteString(i.Travel.DestLocation.Name)
	left.WriteString("\n")
	left.WriteString(style.DimTextStyle.Render(i.Travel.DestLocation.Continent.Name))
	left.WriteString("\n")
	left.WriteString(fmt.Sprintf("%s | %s",
		style.DimTextStyle.Render(i.Travel.DestLocation.Type.Name),
		style.LocationBiomeStyle(i.Travel.DestLocation.Biome.Code).
			Render(i.Travel.DestLocation.Biome.Name)))

	rightBottom := ""

	if i.Travel.Price > 0 {
		rightBottom += fmt.Sprintf("Price : %d", i.Travel.Price)
	}

	if i.Travel.RequiredLocationFeature.Name != "" {
		if len(rightBottom) > 0 {
			rightBottom += "\n"
		}
		rightBottom += fmt.Sprintf("%s", i.Travel.RequiredLocationFeature.Name)
	}

	right.WriteString(helper.HoursDecFormat(i.Travel.Duration))
	right.WriteString(rightBottom)

	tui := s.Width(m.Width() - 2).Render(
		lipgloss.JoinHorizontal(lipgloss.Center,
			style.TextStyle.Width(width/2).
				AlignHorizontal(lipgloss.Left).
				Render(left.String()),
			style.TextStyle.Width(width/2).
				AlignHorizontal(lipgloss.Right).
				Render(right.String())))

	fmt.Fprint(w, tui)
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
