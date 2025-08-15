package skillssummary

import (
	"farental/core/data/api"
	"farental/internal/keybind"
	"farental/internal/orvyn"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"strconv"
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

	viewport viewport.Model

	skills []api.CharacterSkillResponse

	contentSize orvyn.Size
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.title = lokyn.L("Skills")

	w.viewport = viewport.New(0, 0)

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Up):
			w.viewport.ScrollUp(1)
		case key.Matches(msg, keybind.Down):
			w.viewport.ScrollDown(1)
		}
	}

	return nil
}

func (w *Widget) Render() string {
	w.refresh()

	content := lipgloss.JoinVertical(lipgloss.Left,
		style.DimUnderlinedTitleStyle.
			Width(w.contentSize.Width).
			Render(w.title),
		w.viewport.View())

	return style.BlurredStyle.
		Width(w.contentSize.Width).
		Height(w.contentSize.Height).
		Render(content)
}

func (w *Widget) Resize(size orvyn.Size) {
	s := size

	size.Width -= style.BlurredStyle.GetHorizontalFrameSize()
	size.Height -= style.BlurredStyle.GetVerticalFrameSize()

	w.contentSize = size
	w.viewport.Width = size.Width
	w.viewport.Height = size.Height -
		lipgloss.Height(style.DimUnderlinedTitleStyle.Render(" "))

	if !orvyn.SameSize(s, w.GetSize()) {
		w.refresh()
	}

	w.BaseWidget.Resize(s)
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(15, 5)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(30, 17)
}

func (w *Widget) renderSkill(skill api.CharacterSkillResponse, addReturn bool, column *column) {
	column.skillStr.WriteString(style.NormalStyle.Render(skill.Name))
	column.expStr.WriteString(style.NeutralLessDimTextStyle.
		Render(fmt.Sprintf("(%d / %d)",
			skill.CurrentXp, skill.NextLevelXp)))
	column.lvlStr.WriteString(style.NormalStyle.
		Render(fmt.Sprintf("%s %s", lokyn.L("lvl."),
			style.HighlightStyle.Render(strconv.Itoa(skill.Level)))))

	if addReturn {
		column.addReturn()
	}
}

func (w *Widget) UpdateData(characterInfo *api.CharacterInfoResponse) {
	w.skills = characterInfo.Skills
}

func (w *Widget) refresh() {
	var col column
	var addReturn bool

	addReturn = true

	for i, skill := range w.skills {
		if i == len(w.skills)-1 {
			addReturn = false
		}

		w.renderSkill(skill, addReturn, &col)
	}

	w.viewport.SetContent(col.render(w.contentSize.Width))
}
