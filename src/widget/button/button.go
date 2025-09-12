package button

import (
	"farental/internal/keybind"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	label string

	contentSize orvyn.Size

	OnFocusCallback   func()
	OnBlurCallback    func()
	OnClickedCallback func() tea.Cmd

	style lipgloss.Style
}

func New(label string) *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

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

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height -= w.style.GetVerticalFrameSize()

	w.contentSize = size
}

func (w *Widget) Render() string {
	return w.style.Width(w.contentSize.Width).
		Height(w.contentSize.Height).
		Render(w.label)
}

func (w *Widget) OnFocus() {
	w.style = orvyn.GetTheme().Style(theme.FocusedWidgetStyleID)

	if w.OnFocusCallback != nil {
		w.OnFocusCallback()
	}
}

func (w *Widget) OnBlur() {
	w.style = orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)

	if w.OnBlurCallback != nil {
		w.OnBlurCallback()
	}
}

func (w *Widget) OnEnterInput() {
}

func (w *Widget) OnExitInput() {
}

func (w *Widget) SetLabel(label string) {
	w.label = label
}
