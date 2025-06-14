package charactervitalinfo

import (
	"farental/core/data/api"
	"farental/internal/lang"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

var (
	styleParagraph = lipgloss.NewStyle().AlignHorizontal(lipgloss.Center)
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
	Money          int

	RaceStyle lipgloss.Style

	HpBar progress.Model
	MpBar progress.Model
}

func New(width int) Model {
	m := Model{
		HpBar: progress.New(progress.WithSolidFill("#c90000")),
		MpBar: progress.New(progress.WithSolidFill("#272de8")),
	}

	barWidth := (width / 3) - 2

	m.HpBar.ShowPercentage = false
	m.HpBar.Width = barWidth
	m.MpBar.ShowPercentage = false
	m.MpBar.Width = barWidth

	styleParagraph = styleParagraph.Width(width / 3)

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
	left.WriteString(strings.Repeat(fmt.Sprintf("\n%s", m.HpBar.ViewAs(m.HpPercent)), 3))

	center.WriteString(m.FullName)
	center.WriteString("\n")
	center.WriteString(m.RaceStyle.Render(m.RaceName))
	center.WriteString("\n")
	center.WriteString(style.TextStyle.Foreground(
		lipgloss.Color(style.ColorHighlight)).
		Render(fmt.Sprintf("%d Ǥ", m.Money)))
	center.WriteString("\n")
	center.WriteString(style.TextStyle.Foreground(
		lipgloss.Color(style.ColorSpecialHighlight)).
		Render(fmt.Sprintf("%s : %d", lang.L("Power"), m.Power)))

	right.WriteString(fmt.Sprintf("MP (%d/%d)",
		m.MpCurrentValue, m.MpMaxValue))
	right.WriteString(strings.Repeat(fmt.Sprintf("\n%s", m.MpBar.ViewAs(m.MpPercent)), 3))

	return lipgloss.JoinHorizontal(lipgloss.Top,
		styleParagraph.Render(left.String()),
		styleParagraph.Render(center.String()),
		styleParagraph.Render(right.String()))
}

func (m *Model) UpdateData(characterInfo *api.CharacterInfoResponse, money int) {
	m.FullName = style.BoldTextStyle.Render(
		fmt.Sprintf("%s %s", characterInfo.FirstName, characterInfo.LastName))
	m.RaceName = characterInfo.RaceName
	m.RaceStyle = style.RaceStyle(m.RaceName)
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

	m.Money = money
}
