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

func New() Model {
	m := Model{
		VPDescription: viewport.New(80, 5),
	}

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

	b.WriteString(m.LocationName)
	b.WriteString("\n")
	b.WriteString(m.ContinentName)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%s | %s", m.LocationType, m.LocationBiome))
	b.WriteString("\n")
	b.WriteString(m.VPDescription.View())

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
		Width(m.VPDescription.Width).
		Render(m.LocationDescription))
}
