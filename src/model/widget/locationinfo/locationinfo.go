package locationinfo

import (
	"farental/core/data/api"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"sort"
	"strings"
)

var (
	styleCenterContent = lipgloss.NewStyle().AlignHorizontal(lipgloss.Center)
	styleBottomBorder  = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(style.ColorHighlightDim)).
		BorderTop(false).BorderRight(false).BorderLeft(false)
)

type Model struct {
	LocationName        string
	ContinentName       string
	LocationType        string
	LocationBiome       string
	LocationDescription string
	LocationFeatures    []string

	BiomeStyle lipgloss.Style

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
	var top strings.Builder
	var tui strings.Builder

	top.WriteString(styleCenterContent.
		Render(style.BoldTextStyle.Render(m.LocationName)))
	top.WriteString("\n")
	top.WriteString(styleCenterContent.
		Render(style.NeutralDimTextStyle.Render(m.ContinentName)))
	top.WriteString("\n")
	top.WriteString(styleCenterContent.
		Render(fmt.Sprintf("%s | %s",
			style.NeutralDimTextStyle.Render(m.LocationType),
			m.BiomeStyle.Italic(true).Render(m.LocationBiome))))

	if len(m.LocationFeatures) > 0 {
		top.WriteString("\n")

		str := ""

		for i, f := range m.LocationFeatures {
			if i > 0 {
				str += " • "
			}

			str += f
		}

		top.WriteString(styleCenterContent.Render(style.DimTextStyle.Render(str)))
	}

	tui.WriteString(styleBottomBorder.Render(top.String()))
	tui.WriteString("\n")
	tui.WriteString(styleCenterContent.
		Render(m.VPDescription.View()))

	if m.VPDescription.TotalLineCount() > m.VPDescription.VisibleLineCount() {
		tui.WriteString("V")
	}

	return tui.String()
}

func (m *Model) UpdateData(locationInfo *api.LocationResponse) {
	m.LocationName = locationInfo.Name
	m.ContinentName = locationInfo.Continent.Name
	m.LocationType = locationInfo.Type.Name
	m.LocationBiome = locationInfo.Biome.Name
	m.BiomeStyle = style.LocationBiomeStyle(locationInfo.Biome.Code)
	m.LocationDescription = locationInfo.Description

	m.LocationFeatures = make([]string, 0)

	if len(locationInfo.Features) > 0 {
		sort.Slice(locationInfo.Features, func(i, j int) bool {
			return locationInfo.Features[i].Name < locationInfo.Features[j].Name
		})

		for _, f := range locationInfo.Features {
			if !f.IsAction {
				continue
			}

			m.LocationFeatures = append(m.LocationFeatures, f.Name)
		}
	}

	// Set the Width before the render to wrap text
	m.VPDescription.SetContent(styleCenterContent.
		Render(m.LocationDescription))
}
