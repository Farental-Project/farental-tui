package multivalueselector

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

type Value interface {
	RenderValue() string
}

type Style struct {
	FocusedWidget  lipgloss.Style
	BlurredWidget  lipgloss.Style
	BlurredControl lipgloss.Style
	FocusedControl lipgloss.Style
	BlurredValue   lipgloss.Style
	FocusedValue   lipgloss.Style
}

type Keybind struct {
	Next     key.Binding
	Previous key.Binding
}

type Widget[T Value] struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	Style   Style
	Keybind Keybind

	values map[string]T
	keys   []string

	selectedIndex int

	widgetStyle  lipgloss.Style
	controlStyle lipgloss.Style
	valueStyle   lipgloss.Style

	contentSize orvyn.Size

	Looping bool
}

func New[T Value]() *Widget[T] {
	w := new(Widget[T])

	t := orvyn.GetTheme()
	dts := t.Style(theme.DimTextStyleID)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.values = make(map[string]T)
	w.keys = make([]string, 0)

	w.selectedIndex = 0

	w.Style = Style{
		FocusedWidget:  t.Style(theme.FocusedWidgetStyleID),
		BlurredWidget:  t.Style(theme.BlurredWidgetStyleID),
		BlurredControl: dts,
		FocusedControl: t.Style(theme.HighlightTextStyleID),
		BlurredValue:   dts,
		FocusedValue:   t.Style(theme.NormalTextStyleID),
	}

	w.Keybind = Keybind{
		Next: key.NewBinding(
			key.WithKeys("right"),
		),
		Previous: key.NewBinding(
			key.WithKeys("left"),
		),
	}

	w.widgetStyle = w.Style.BlurredWidget
	w.controlStyle = w.Style.BlurredControl
	w.valueStyle = w.Style.BlurredValue

	w.Looping = false

	w.OnBlur()

	return w
}

func (w *Widget[T]) OnFocus() {
	w.widgetStyle = w.Style.FocusedWidget
	w.controlStyle = w.Style.FocusedControl
	w.valueStyle = w.Style.FocusedValue
}

func (w *Widget[T]) OnBlur() {
	w.widgetStyle = w.Style.BlurredWidget
	w.controlStyle = w.Style.BlurredControl
	w.valueStyle = w.Style.BlurredValue
}

func (w *Widget[T]) OnEnterInput() {}

func (w *Widget[T]) OnExitInput() {}

func (w *Widget[T]) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, w.Keybind.Next):
			w.selectedIndex++

			if w.selectedIndex > len(w.values)-1 {
				if w.Looping {
					w.selectedIndex = 0
				} else {
					w.selectedIndex = len(w.values) - 1
				}
			}

		case key.Matches(msg, w.Keybind.Previous):
			w.selectedIndex--

			if w.selectedIndex < 0 {
				if w.Looping {
					w.selectedIndex = len(w.values) - 1
				} else {
					w.selectedIndex = 0
				}
			}

		}
	}

	return nil
}

func (w *Widget[T]) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	size.Width -= w.widgetStyle.GetHorizontalFrameSize()
	size.Height -= w.widgetStyle.GetVerticalFrameSize()

	w.contentSize = size
}

func (w *Widget[T]) Render() string {
	var b strings.Builder
	var margin int

	size := w.contentSize

	margin += 4 // "< " & " >"

	b.WriteString(w.controlStyle.Render("< "))
	b.WriteString(w.valueStyle.Width(size.Width - margin).
		AlignHorizontal(lipgloss.Center).
		Render(w.GetSelectedValue().RenderValue()))
	b.WriteString(w.controlStyle.Render(" >"))

	return w.widgetStyle.
		Width(size.Width).
		AlignHorizontal(lipgloss.Center).
		Render(b.String())
}

func (w *Widget[T]) SetValues(keys []string, values map[string]T) {
	w.keys = keys
	w.values = values
}

func (w *Widget[T]) GetSelectedValue() T {
	var empty T

	if len(w.keys) == 0 {
		return empty
	}

	return w.values[w.keys[w.selectedIndex]]
}

func (w *Widget[T]) SetSelected(index int) {
	if index < 0 || index >= len(w.values) {
		return
	}

	w.selectedIndex = index
}
