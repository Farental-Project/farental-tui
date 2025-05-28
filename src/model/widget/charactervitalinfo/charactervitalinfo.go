package charactervitalinfo

import (
	"farental/core/data/api"
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

var (
	styleParagraph = lipgloss.NewStyle().Margin(1, 2)
	styleBorder    = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#39d800"))
)

type Model struct {
	FullName       string
	RaceName       string
	Power          int
	HpMaxValue     int
	HpCurrentValue int
	HpPercent      float64
	MpMaxValue     int
	MpCurrentValue int
	MpPercent      float64

	HpBar progress.Model
	MpBar progress.Model
}

func New() Model {
	m := Model{
		HpBar: progress.New(progress.WithSolidFill("#c90000")),
		MpBar: progress.New(progress.WithSolidFill("#272de8")),
	}

	m.HpBar.ShowPercentage = false
	m.HpBar.Width = 20
	m.MpBar.ShowPercentage = false
	m.MpBar.Width = 20

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	var left strings.Builder
	var center strings.Builder
	var right strings.Builder

	left.WriteString(fmt.Sprintf("HP (%d/%d)",
		m.HpCurrentValue, m.HpMaxValue))
	left.WriteString("\n")
	left.WriteString(m.HpBar.ViewAs(m.HpPercent))
	left.WriteString("\n")
	left.WriteString(m.HpBar.ViewAs(m.HpPercent))

	width := lipgloss.Width(m.FullName)
	center.WriteString(m.FullName)
	center.WriteString("\n")
	center.WriteString(lipgloss.PlaceHorizontal(
		width, lipgloss.Center, m.RaceName))
	center.WriteString("\n")
	center.WriteString(lipgloss.PlaceHorizontal(
		width, lipgloss.Center, fmt.Sprintf("%d", m.Power)))

	right.WriteString(fmt.Sprintf("MP (%d/%d)",
		m.MpCurrentValue, m.MpMaxValue))
	right.WriteString("\n")
	right.WriteString(m.MpBar.ViewAs(m.MpPercent))
	right.WriteString("\n")
	right.WriteString(m.MpBar.ViewAs(m.MpPercent))

	return styleBorder.Render(lipgloss.JoinHorizontal(lipgloss.Center,
		styleParagraph.Render(left.String()),
		styleParagraph.Render(center.String()),
		styleParagraph.Render(right.String())))
}

func (m *Model) UpdateData(characterInfo *api.CharacterInfoResponse) {
	m.FullName = fmt.Sprintf("%s %s", characterInfo.FirstName, characterInfo.LastName)
	m.RaceName = characterInfo.RaceName
	m.Power = characterInfo.Power

	for _, stat := range characterInfo.Stats {
		if stat.Code == "hp" {
			m.HpMaxValue = stat.MaxValue
			m.HpCurrentValue = stat.Value
			continue
		}

		if stat.Code == "mp" {
			m.MpMaxValue = stat.MaxValue
			m.MpCurrentValue = stat.Value
			continue
		}
	}

	percent := 100 * m.HpCurrentValue / m.HpMaxValue
	m.HpPercent = float64(percent) / 100

	percent = 100 * m.MpCurrentValue / m.MpMaxValue
	m.MpPercent = float64(percent) / 100
}
