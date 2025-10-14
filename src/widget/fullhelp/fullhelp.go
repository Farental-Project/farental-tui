package fullhelp

import (
	ftheme "farental/internal/theme"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

type Style struct {
	Title lipgloss.Style
	Help  lipgloss.Style
}

type Widget struct {
	orvyn.BaseWidget

	Style Style

	title string

	titleHeight int
}

func New() *Widget {
	w := new(Widget)
	t := orvyn.GetTheme()

	w.title = lokyn.L("Help")

	w.BaseWidget = orvyn.NewBaseWidget()

	w.Style = Style{
		Title: t.Style(ftheme.DimUnderlinedTextStyleID),
		Help: t.Style(theme.NormalTextStyleID).
			Align(lipgloss.Center, lipgloss.Center),
	}

	w.titleHeight = lipgloss.Height(w.Style.Title.Render(w.title))

	return w
}

func (w *Widget) Render() string {
	var b strings.Builder

	contentSize := w.GetContentSize()

	w.Style.Title = w.Style.Title.Width(contentSize.Width)

	b.WriteString(w.Style.Title.
		Render(w.title))
	b.WriteString("\n")
	b.WriteString(w.Style.Help.
		Width(contentSize.Width).
		Height(contentSize.Height).Render(
		bubblehelp.ViewAll(
			bubblehelp.GetCurrentContextKeymap(),
			w.GetSize().Width)))

	return w.GetStyle().
		Width(contentSize.Width).
		Height(contentSize.Height).
		Render(b.String())
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(10, 1)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.GetRenderSize(lipgloss.NewStyle(),
		bubblehelp.ViewAll(bubblehelp.GetCurrentContextKeymap(), 50))
}

func (w *Widget) SetTitle(title string) {
	w.title = title
}
