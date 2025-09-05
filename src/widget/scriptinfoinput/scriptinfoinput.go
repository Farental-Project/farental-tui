package scriptinfoinput

import (
	"farental/internal/keybind"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/checkbox"
	"github.com/halsten-dev/orvyn/widget/textarea"
	"github.com/halsten-dev/orvyn/widget/textinput"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	title       *orvyn.SimpleRenderable
	name        *textinput.Widget
	description *textarea.Widget
	private     *checkbox.Widget

	focusManager *orvyn.FocusManager

	layout *layout.VBoxLayout

	style lipgloss.Style

	contentSize orvyn.Size
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.title = orvyn.NewSimpleRenderable(lokyn.L("Script information"))
	w.title.Style = orvyn.GetTheme().Style(theme.TitleStyleID)

	w.name = textinput.New()
	w.name.Placeholder = lokyn.L("Script name")

	w.description = textarea.New()
	w.description.Placeholder = lokyn.L("Script description")
	w.description.ShowLineNumbers = false

	w.private = checkbox.New(lokyn.L("Private"))

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.name)
	w.focusManager.Add(w.description)
	w.focusManager.Add(w.private)

	w.layout = layout.NewMaxWidthVBoxLayout(
		0, []orvyn.Renderable{
			w.title,
			w.name,
			w.description,
			w.private,
		},
	)

	w.OnBlur()

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(m, keybind.Esc):
			orvyn.SwitchToPreviousScreen()
		}
	}

	cmd := w.focusManager.Update(msg)

	return cmd
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height -= w.style.GetVerticalFrameSize()

	w.layout.Resize(size)
	w.contentSize = size
}

func (w *Widget) Render() string {
	return w.style.
		Width(w.contentSize.Width).
		Height(w.contentSize.Height).
		Render(w.layout.Render())
}

func (w *Widget) OnFocus() {
	w.style = orvyn.GetTheme().Style(theme.FocusedWidgetStyleID)
}

func (w *Widget) OnBlur() {
	w.style = orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)
}

func (w *Widget) OnEnterInput() {}

func (w *Widget) OnExitInput() {}
