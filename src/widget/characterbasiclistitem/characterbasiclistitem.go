package characterbasiclistitem

import (
	"farental/core/data/api"
	"farental/internal/style"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	style lipgloss.Style

	data *api.CharacterBasicResponse

	contentSize orvyn.Size
}

func Constructor(data *api.CharacterBasicResponse) list.IListItem {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.data = data

	w.OnBlur()

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 5

	w.BaseWidget.Resize(size)

	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height -= w.style.GetVerticalFrameSize()

	w.contentSize = size
}

func (w *Widget) Render() string {
	t := orvyn.GetTheme()

	str := w.style.Width(w.contentSize.Width).
		Height(w.contentSize.Height).Render(
		fmt.Sprintf("%s\n%s\n%s",
			t.Style(theme.HighlightTextStyleID).Render(fmt.Sprintf("%s %s", w.data.FirstName, w.data.LastName)),
			style.RaceStyle(w.data.RaceName).Render(w.data.RaceName),
			t.Style(theme.DimTextStyleID).Render(w.data.LocationName),
		),
	)

	return str
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
