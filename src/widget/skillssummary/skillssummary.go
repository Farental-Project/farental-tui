package skillssummary

import (
	"farental/core/data/api"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/style"
	"fmt"
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

type Widget struct {
	orvyn.BaseWidget

	title string

	skills []api.CharacterSkillResponse

	contentSize orvyn.Size
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.title = lang.L("Skills")

	return w
}

func (w *Widget) Render() string {
	var col column
	var addReturn bool

	addReturn = true

	for i, skill := range w.skills {
		if i == len(w.skills)-1 {
			addReturn = false
		}

		w.renderSkill(skill, addReturn, &col)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		style.DimUnderlinedTitleStyle.
			Width(w.contentSize.Width).
			Render(w.title),
		col.render(w.contentSize.Width))

	return style.BlurredStyle.
		Width(w.contentSize.Width).
		Height(w.contentSize.Height).
		Render(content)
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	size.Width -= style.BlurredStyle.GetHorizontalFrameSize()
	size.Height -= style.BlurredStyle.GetVerticalFrameSize()

	w.contentSize = size
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(15, 5)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(30, 17)
}

func (w *Widget) renderSkill(skill api.CharacterSkillResponse, addReturn bool, column *column) {
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

func (w *Widget) UpdateData(characterInfo *api.CharacterInfoResponse) {
	w.skills = characterInfo.Skills
}
