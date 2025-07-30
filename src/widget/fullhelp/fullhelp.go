package fullhelp

import (
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"strings"
)

type Style struct {
	Widget lipgloss.Style
	Title  lipgloss.Style
	help   lipgloss.Style
}

type Widget struct {
	orvyn.BaseWidget

	Style Style

	title string

	titleHeight int

	helpContentSize orvyn.Size
}

func New() *Widget {
	w := new(Widget)

	w.title = lang.L("Help")

	w.BaseWidget = orvyn.NewBaseWidget()
	
	w.Style = Style{
		Widget: style.BlurredStyle,
		Title:  style.DimUnderlinedTitleStyle,
		help: style.NormalStyle.
			Align(lipgloss.Center, lipgloss.Center),
	}

	w.titleHeight = lipgloss.Height(w.Style.Title.Render(w.title))

	return w
}

func (w *Widget) Render() string {
	var b strings.Builder

	size := w.GetSize()

	b.WriteString(w.Style.Title.
		Render(w.title))
	b.WriteString("\n")
	b.WriteString(w.Style.help.
		Width(w.helpContentSize.Width).
		Height(w.helpContentSize.Height).Render(
		bubblehelp.ViewAll(
			bubblehelp.GetCurrentContextKeymap(),
			w.GetSize().Width)))

	return w.Style.Widget.
		Width(size.Width).
		Height(size.Height).
		Render(b.String())
}

func (w *Widget) Resize(size orvyn.Size) {
	var marginW, marginH int

	marginW += w.Style.Widget.GetBorderLeftSize()
	marginW += w.Style.Widget.GetBorderRightSize()

	marginH += w.Style.Widget.GetBorderTopSize()

	w.Style.Title = w.Style.Title.Width(size.Width - marginW)
	w.helpContentSize.Width = size.Width - marginW
	w.helpContentSize.Height = size.Height - w.titleHeight - marginH
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(10, 1)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.GetRenderSize(lipgloss.NewStyle(),
		bubblehelp.ViewAll(bubblehelp.GetCurrentContextKeymap(), 50))
}
