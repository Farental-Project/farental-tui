package scriptrulelistitem

import (
	"farental/core/data/api"
	"farental/internal/keybind"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	style lipgloss.Style
}

func Constructor(data *api.ScriptRuleBody) list.IListItem {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

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
