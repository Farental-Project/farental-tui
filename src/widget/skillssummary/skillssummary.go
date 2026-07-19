package skillssummary

import (
	"farental/core/data/api"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

type column struct {
	skillStr strings.Builder
	expStr   strings.Builder
	lvlStr   strings.Builder
}

func (c *column) addReturn() {
	c.skillStr.WriteString("\n")
	c.expStr.WriteString("\n")
	c.lvlStr.WriteString("\n")
}

func (c *column) render(width int) string {
	var colWidth int

	style := lipgloss.NewStyle()

	colWidth = width / 3

	nameCol := style.Width(colWidth).
		AlignHorizontal(lipgloss.Left).Render(c.skillStr.String())

	expCol := style.Width(colWidth).
		AlignHorizontal(lipgloss.Center).Render(c.expStr.String())

	lvlCol := style.Width(colWidth).
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

	t := orvyn.GetTheme()
	contentSize := w.GetContentSize()

	content := lipgloss.JoinVertical(lipgloss.Left,
		t.Style(ftheme.DimUnderlinedTextStyleID).
			Width(contentSize.Width).
			Render(w.title),
		w.viewport.View())

	return w.GetStyle().
		Width(contentSize.Width).
		Height(contentSize.Height).
		Render(content)
}

func (w *Widget) Resize(size orvyn.Size) {
	s := size
	currentSize := w.GetSize()

	t := orvyn.GetTheme()

	w.BaseWidget.Resize(size)

	contentSize := w.GetContentSize()

	w.viewport.Width = contentSize.Width
	w.viewport.Height = contentSize.Height -
		lipgloss.Height(t.Style(ftheme.DimUnderlinedTextStyleID).Render(" "))

	if s != currentSize {
		w.refresh()
	}
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(15, 5)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(30, 17)
}

func (w *Widget) renderSkill(skill api.CharacterSkillResponse, addReturn bool, column *column) {
	t := orvyn.GetTheme()

	column.skillStr.WriteString(t.Style(theme.NormalTextStyleID).
		Render(fmt.Sprintf("  %s", skill.Name)))
	column.expStr.WriteString(t.Style(theme.NeutralTextStyleID).
		Render(fmt.Sprintf("(%d / %d)",
			skill.CurrentXp, skill.NextLevelXp)))
	column.lvlStr.WriteString(t.Style(theme.NormalTextStyleID).
		Render(fmt.Sprintf("%s %s", lokyn.L("lvl."),
			t.Style(theme.HighlightTextStyleID).Render(strconv.Itoa(skill.Level)))))

	if addReturn {
		column.addReturn()
	}
}

func (w *Widget) renderCategory(category string, column *column) {
	ts := orvyn.GetTheme().Style(theme.TitleStyleID)

	column.skillStr.WriteString(ts.Render(category))
	column.expStr.WriteString("")
	column.lvlStr.WriteString("")

	column.addReturn()
}

func (w *Widget) UpdateData(skills []api.CharacterSkillResponse) {
	w.skills = skills
}

func (w *Widget) refresh() {
	var col column
	var addReturn bool
	var craftCatAdded bool

	addReturn = true
	craftCatAdded = false

	for i, skill := range w.skills {
		if i == 0 {
			w.renderCategory(lokyn.L("Fighting"), &col)
		}

		if !skill.IsFightingSkill && !craftCatAdded {
			col.addReturn()
			w.renderCategory(lokyn.L("Craftsmanship"), &col)
			craftCatAdded = true
		}

		if i == len(w.skills)-1 {
			addReturn = false
		}

		w.renderSkill(skill, addReturn, &col)
	}

	w.viewport.SetContent(col.render(w.GetContentSize().Width))
}
