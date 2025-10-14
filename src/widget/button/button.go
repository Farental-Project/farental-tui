package button

import (
	"farental/internal/keybind"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/orvyn"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	label string

	OnFocusCallback   func()
	OnBlurCallback    func()
	OnClickedCallback func() tea.Cmd
}

func New(label string) *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.label = label

	w.OnBlur()

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(m, keybind.Space):
			if w.IsFocused() {
				if w.OnClickedCallback != nil {
					return w.OnClickedCallback()
				}
			}
		}
	}

	return nil
}

func (w *Widget) Render() string {
	contentSize := w.GetContentSize()

	return w.GetStyle().Width(contentSize.Width).
		Height(contentSize.Height).
		Render(w.label)
}

func (w *Widget) OnFocus() {
	w.BaseFocusable.OnFocus()

	if w.OnFocusCallback != nil {
		w.OnFocusCallback()
	}
}

func (w *Widget) OnBlur() {
	w.BaseFocusable.OnBlur()

	if w.OnBlurCallback != nil {
		w.OnBlurCallback()
	}
}

func (w *Widget) SetLabel(label string) {
	w.label = label
}
