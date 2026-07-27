package statssummary

import (
	"farental/core/data"
	"farental/core/data/api"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
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

	rightPart := lipgloss.NewStyle().Width(rightWidth).
		AlignHorizontal(lipgloss.Right).Render(c.valStr.String())

	return lipgloss.JoinHorizontal(lipgloss.Center,
		leftPart,
		rightPart)
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	title string

	viewport viewport.Model

	arrowStyle lipgloss.Style

	statMap data.StatMap
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.title = lokyn.L("Stats")

	w.viewport = viewport.New(0, 0)

	w.OnBlur()

	return w
}

func (w *Widget) OnFocus() {
	w.BaseFocusable.OnFocus()

	w.arrowStyle = orvyn.GetTheme().Style(theme.NormalTextStyleID)
}

func (w *Widget) OnBlur() {
	w.BaseFocusable.OnBlur()

	w.arrowStyle = orvyn.GetTheme().Style(theme.DimTextStyleID)
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
		w.renderViewport())

	return w.GetStyle().
		Width(contentSize.Width).
		Height(contentSize.Height).
		Render(content)
}

// renderViewport renders the viewport and overlays scroll indicators on the
// last column: an up arrow on the first line when there is content above, and
// a down arrow on the last line when there is content below.
func (w *Widget) renderViewport() string {
	view := w.viewport.View()

	if w.viewport.Width < 1 || w.viewport.Height < 1 {
		return view
	}

	return helper.OverlayScrollArrows(view, w.viewport.Width, w.arrowStyle,
		!w.viewport.AtTop(), !w.viewport.AtBottom())
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

func (w *Widget) renderStat(statCode data.StatCode, addReturn bool, column *column) {
	s := w.statMap[statCode]

	column.statStr.WriteString(orvyn.GetTheme().Style(theme.NormalTextStyleID).Render(s.Name))
	column.valStr.WriteString(lipgloss.NewStyle().
		Render(strconv.Itoa(s.Value)))

	if addReturn {
		column.addReturn()
	}
}

func (w *Widget) GotoTop() {
	w.viewport.GotoTop()
}

func (w *Widget) UpdateData(stats []api.CharacterStatResponse) {
	w.statMap = data.NewStatMap(stats)
}

func (w *Widget) refresh() {
	var col column

	w.renderStat(data.INIStat, true, &col)
	w.renderStat(data.STRStat, true, &col)
	w.renderStat(data.INTStat, true, &col)
	w.renderStat(data.LUKStat, true, &col)
	w.renderStat(data.AGIStat, true, &col)
	w.renderStat(data.DEFStat, true, &col)
	w.renderStat(data.MDEStat, true, &col)
	w.renderStat(data.ATKStat, false, &col)

	helper.SetScrollableContent(&w.viewport, w.GetContentSize().Width, col.render)
}
