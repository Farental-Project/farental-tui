package characterbasiclistitem

import (
	"farental/core/data/api"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	name     *orvyn.SimpleRenderable
	race     *orvyn.SimpleRenderable
	location *orvyn.SimpleRenderable

	style lipgloss.Style

	data *api.CharacterBasicResponse

	layout *layout.CenterLayout
}

func Constructor(data *api.CharacterBasicResponse) list.IListItem {
	w := new(Widget)

	t := orvyn.GetTheme()

	w.BaseWidget = orvyn.NewBaseWidget()

	w.data = data

	w.name = orvyn.NewSimpleRenderable("")
	w.name.Style = t.Style(theme.TitleStyleID)

	w.race = orvyn.NewSimpleRenderable("")
	// TODO: Manage race colors
	w.race.Style = t.Style(theme.HighlightTextStyleID)

	w.location = orvyn.NewSimpleRenderable("")
	w.location.Style = t.Style(theme.DimSecondaryTextStyleID)

	w.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxLayout(0,
			[]orvyn.Renderable{
				w.name,
				w.race,
				w.location,
			},
		),
	)

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height -= w.style.GetVerticalFrameSize()

	w.layout.Resize(size)
}

func (w *Widget) Render() string {
	return w.layout.Render()
}

func (w *Widget) OnFocus() {
	w.style = orvyn.GetTheme().Style(theme.FocusedWidgetStyleID)
}

func (w *Widget) OnBlur() {
	w.style = orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)
}

func (w *Widget) OnEnterInput() {}

func (w *Widget) OnExitInput() {}

func (w *Widget) FilterValue() string {
	return ""
}
