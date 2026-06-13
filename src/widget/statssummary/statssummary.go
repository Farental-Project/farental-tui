package statssummary

import (
	"farental/core/data"
	"farental/core/data/api"
	ftheme "farental/internal/theme"
	"strconv"
	"strings"

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

	title string

	statMap data.StatMap
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.title = lokyn.L("Stats")

	return w
}

func (w *Widget) Render() string {
	var col column

	t := orvyn.GetTheme()

	w.renderStat(data.INIStat, true, &col)
	w.renderStat(data.STRStat, true, &col)
	w.renderStat(data.INTStat, true, &col)
	w.renderStat(data.LUKStat, true, &col)
	w.renderStat(data.AGIStat, true, &col)
	w.renderStat(data.DEFStat, true, &col)
	w.renderStat(data.MDEStat, true, &col)
	w.renderStat(data.ATKStat, false, &col)

	contentSize := w.GetContentSize()

	content := lipgloss.JoinVertical(lipgloss.Left,
		t.Style(ftheme.DimUnderlinedTextStyleID).
			Width(contentSize.Width).
			Render(w.title),
		col.render(contentSize.Width))

	return w.GetStyle().
		Width(contentSize.Width).
		Height(contentSize.Height).
		Render(content)
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
	// column.sepStr.WriteString(style.DimTextStyle.Render(" : "))
	column.valStr.WriteString(lipgloss.NewStyle().
		Render(strconv.Itoa(s.Value)))

	if addReturn {
		column.addReturn()
	}
}

func (w *Widget) UpdateData(characterInfo *api.CharacterInfoResponse) {
	w.statMap = data.NewStatMap(characterInfo.Stats)
}
