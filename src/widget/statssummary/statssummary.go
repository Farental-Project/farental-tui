package statssummary

import (
	"farental/core/data"
	"farental/core/data/api"
	"farental/internal/orvyn"
	"farental/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
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

type Widget struct {
	orvyn.BaseWidget

	title string

	statMap data.StatMap

	contentSize orvyn.Size
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.title = lokyn.L("Stats")

	return w
}

func (w *Widget) Render() string {
	var col column

	w.renderStat(data.INIStat, true, &col)
	w.renderStat(data.STRStat, true, &col)
	w.renderStat(data.INTStat, true, &col)
	w.renderStat(data.LUKStat, true, &col)
	w.renderStat(data.PREStat, true, &col)
	w.renderStat(data.AGIStat, true, &col)
	w.renderStat(data.DEFStat, true, &col)
	w.renderStat(data.MDEStat, true, &col)
	w.renderStat(data.ATKStat, false, &col)

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

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(15, 5)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(30, 17)
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	size.Width -= style.BlurredStyle.GetHorizontalFrameSize()
	size.Height -= style.BlurredStyle.GetVerticalFrameSize()

	w.contentSize = size
}

func (w *Widget) renderStat(statCode data.StatCode, addReturn bool, column *column) {
	s := w.statMap[statCode]

	column.statStr.WriteString(style.NormalStyle.Render(s.Name))
	// column.sepStr.WriteString(style.DimTextStyle.Render(" : "))
	column.valStr.WriteString(style.TextStyle.
		Render(strconv.Itoa(s.Value)))

	if addReturn {
		column.addReturn()
	}
}

func (w *Widget) UpdateData(characterInfo *api.CharacterInfoResponse) {
	w.statMap = data.NewStatMap(characterInfo.Stats)
}
