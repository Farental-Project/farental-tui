package locationinfo

import (
	"farental/core/data/api"
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

var (
	styleCenterContent = lipgloss.NewStyle().AlignHorizontal(lipgloss.Center)
)

type Model struct {
	LocationName        string
	ContinentName       string
	LocationType        string
	LocationBiome       string
	LocationDescription string

	VPDescription viewport.Model
}

func New(width int) Model {
	m := Model{
		VPDescription: viewport.New(width, 5),
	}

	styleCenterContent = styleCenterContent.Width(width)

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(styleCenterContent.
		Render(m.LocationName))
	b.WriteString("\n")
	b.WriteString(styleCenterContent.
		Render(m.ContinentName))
	b.WriteString("\n")
	b.WriteString(styleCenterContent.
		Render(fmt.Sprintf("%s | %s",
			m.LocationType, m.LocationBiome)))
	b.WriteString("\n")
	b.WriteString(styleCenterContent.
		Render(m.VPDescription.View()))

	if m.VPDescription.TotalLineCount() > m.VPDescription.VisibleLineCount() {
		b.WriteString("V")
	}

	return b.String()
}

func (m *Model) UpdateData(locationInfo *api.LocationResponse) {
	m.LocationName = locationInfo.Name
	m.ContinentName = locationInfo.Continent.Name
	m.LocationType = locationInfo.Type.Name
	m.LocationBiome = locationInfo.Biome.Name
	m.LocationDescription = locationInfo.Description
	// Set the Width before the render to wrap text
	m.VPDescription.SetContent(styleCenterContent.
		Render(m.LocationDescription))
}
