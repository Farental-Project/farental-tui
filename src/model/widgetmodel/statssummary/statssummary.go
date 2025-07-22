package statssummary

import (
	"farental/core/data"
	"farental/core/data/api"
	"farental/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strconv"
	"strings"
)

type column struct {
	statStr strings.Builder
	sepStr  strings.Builder
	valStr  strings.Builder
}

func (c *column) reset() {
	c.statStr.Reset()
	c.sepStr.Reset()
	c.valStr.Reset()
}

func (c *column) addReturn() {
	c.statStr.WriteString("\n")
	c.sepStr.WriteString("\n")
	c.valStr.WriteString("\n")
}

func (c *column) render(width int) string {
	leftPart := lipgloss.JoinHorizontal(lipgloss.Left,
		c.statStr.String(),
		c.sepStr.String())

	rightWidth := width - lipgloss.Width(leftPart)

	rightPart := style.TextStyle.Width(rightWidth).
		AlignHorizontal(lipgloss.Right).Render(c.valStr.String())

	return lipgloss.JoinHorizontal(lipgloss.Center,
		leftPart,
		rightPart)
}

type Model struct {
	statMap data.StatMap
	width   int
}

func New(width int) Model {
	m := Model{width: width}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	var col column

	m.renderStat(data.INIStat, true, &col)
	m.renderStat(data.STRStat, true, &col)
	m.renderStat(data.INTStat, true, &col)
	m.renderStat(data.LUKStat, true, &col)
	m.renderStat(data.PREStat, true, &col)
	m.renderStat(data.AGIStat, true, &col)
	m.renderStat(data.DEFStat, true, &col)
	m.renderStat(data.MDEStat, true, &col)
	m.renderStat(data.ATKStat, false, &col)

	return col.render(m.width)
}

func (m Model) renderStat(statCode data.StatCode, addReturn bool, column *column) {
	s := m.statMap[statCode]

	column.statStr.WriteString(style.TitleStyle.Render(s.Name))
	column.sepStr.WriteString(style.TitleStyle.Render(" : "))
	column.valStr.WriteString(style.HighlightStyle.
		Render(strconv.Itoa(s.Value)))

	if addReturn {
		column.addReturn()
	}
}

func (m *Model) UpdateData(characterInfo *api.CharacterInfoResponse) {
	m.statMap = data.NewStatMap(characterInfo.Stats)
}
