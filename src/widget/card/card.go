package card

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

type Widget struct {
	orvyn.BaseWidget

	Title string

	Content string
}

func New(title, content string) *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.Title = title
	w.Content = content

	return w
}

func (w *Widget) Render() string {
	t := orvyn.GetTheme()
	ts := t.Style(theme.TitleStyleID)
	ds := t.Style(theme.DimTextStyleID)

	style := w.GetStyle()
	contentSize := w.GetContentSize()

	title := ts.Width(contentSize.Width).Render(w.Title)
	content := ds.Width(contentSize.Width).Render(w.Content)

	return style.Width(contentSize.Width).Render(
		lipgloss.JoinVertical(lipgloss.Left, title, content))

}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(10, 4)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(20, 4)
}
