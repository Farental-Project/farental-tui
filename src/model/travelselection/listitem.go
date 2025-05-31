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

	if index == m.Index() {
		s = style.FocusedStyle
	} else {
		s = style.BlurredStyle
	}

	left := lipgloss.JoinVertical(
		lipgloss.Left,
		i.Travel.DestLocation.Name,
		i.Travel.DestLocation.Continent.Name,
		fmt.Sprintf("%s | %s",
			i.Travel.DestLocation.Type.Name,
			i.Travel.DestLocation.Biome.Name))

	rightBottom := ""

	if i.Travel.Price > 0 {
		rightBottom += fmt.Sprintf("Price : %d", i.Travel.Price)
	}

	if i.Travel.RequestedLocationFeature.Name != "" {
		if len(rightBottom) > 0 {
			rightBottom += "\n"
		}
		rightBottom += fmt.Sprintf("%s", i.Travel.RequestedLocationFeature.Name)
	}

	right := lipgloss.JoinVertical(
		lipgloss.Right,
		helper.HoursDecFormat(i.Travel.Duration),
		rightBottom)

	tui := s.Width(m.Width() - 2).Render(
		lipgloss.JoinHorizontal(lipgloss.Center, left, right))

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
