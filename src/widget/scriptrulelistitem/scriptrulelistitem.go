package scriptrulelistitem

import (
	cdata "farental/core/data"
	"farental/core/data/api"
	"farental/internal/keybind"
	"farental/widget/multivalueselector"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	mvsTarget *multivalueselector.Widget[cdata.Target]

	layout *layout.HBoxGrowLayout

	style lipgloss.Style
}

func Constructor(data *api.ScriptRuleBody) list.IListItem {
	w := new(Widget)

	t := orvyn.GetTheme()
	dts := t.Style(theme.DimTextStyleID)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.mvsTarget = multivalueselector.New[cdata.Target]()
	w.mvsTarget.SetValues(cdata.TargetKeys, cdata.Targets)
	w.mvsTarget.Style = multivalueselector.Style{
		FocusedWidget:  t.Style(theme.FocusedWidgetStyleID),
		BlurredWidget:  t.Style(theme.BlurredWidgetStyleID),
		BlurredControl: dts,
		FocusedControl: t.Style(theme.HighlightTextStyleID),
		BlurredValue:   dts,
		FocusedValue:   t.Style(theme.NormalTextStyleID),
	}
	w.mvsTarget.OnBlur()

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)
}

func (w *Widget) Render() string {
	return ""
}

func (w *Widget) OnFocus() {
	w.style = orvyn.GetTheme().Style(theme.FocusedWidgetStyleID)
}

func (w *Widget) OnBlur() {
	w.style = orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)
}

func (w *Widget) OnEnterInput() {

}

func (w *Widget) OnExitInput() {

}

func (w *Widget) GetEnterInputKeybind() *key.Binding {
	return &keybind.EKey
}

func (w *Widget) FilterValue() string {
	return ""
}
