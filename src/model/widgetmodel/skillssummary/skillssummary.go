package skillssummary

import (
	"farental/core/data/api"
	"farental/internal/lang"
	"farental/style"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type column struct {
	skillStr strings.Builder
	expStr   strings.Builder
	lvlStr   strings.Builder
}

func (c *column) reset() {
	c.skillStr.Reset()
	c.expStr.Reset()
	c.lvlStr.Reset()
}

func (c *column) addReturn() {
	c.skillStr.WriteString("\n")
	c.expStr.WriteString("\n")
	c.lvlStr.WriteString("\n")
}

func (c *column) render(width int) string {
	var colWidth int

	colWidth = width / 3

	nameCol := style.TextStyle.Width(colWidth).
		AlignHorizontal(lipgloss.Left).Render(c.skillStr.String())

	expCol := style.TextStyle.Width(colWidth).
		AlignHorizontal(lipgloss.Center).Render(c.expStr.String())

	lvlCol := style.TextStyle.Width(colWidth).
		AlignHorizontal(lipgloss.Right).Render(c.lvlStr.String())

	return lipgloss.JoinHorizontal(lipgloss.Center,
		nameCol,
		expCol,
		lvlCol)
}

type Model struct {
	skills []api.CharacterSkillResponse
	width  int
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
	var addReturn bool

	addReturn = true

	for i, skill := range m.skills {
		if i == len(m.skills)-1 {
			addReturn = false
		}

		m.renderSkill(skill, addReturn, &col)
	}

	return col.render(m.width)
}

func (m Model) renderSkill(skill api.CharacterSkillResponse, addReturn bool, column *column) {
	column.skillStr.WriteString(style.TitleStyle.Render(skill.Name))
	column.expStr.WriteString(style.DimTextStyle.
		Render(fmt.Sprintf("(%d / %d)",
			skill.CurrentXp, skill.NextLevelXp)))
	column.lvlStr.WriteString(style.HighlightStyle.
		Render(fmt.Sprintf("%s %d", lang.L("lvl."), skill.Level)))

	if addReturn {
		column.addReturn()
	}
}

func (m *Model) UpdateData(characterInfo *api.CharacterInfoResponse) {
	m.skills = characterInfo.Skills
}
