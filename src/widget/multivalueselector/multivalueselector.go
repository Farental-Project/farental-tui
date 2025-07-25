package multivalueselector

import (
	"farental/internal/orvyn"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
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
	orvyn.BaseFocusable

	Style   Style
	Keybind Keybind

	values map[string]T
	keys   []string

	selectedIndex int

	widgetStyle  lipgloss.Style
	controlStyle lipgloss.Style
	valueStyle   lipgloss.Style

	Looping bool
}

func New[T Value]() *Widget[T] {
	w := new(Widget[T])

	w.values = make(map[string]T)
	w.keys = make([]string, 0)

	w.selectedIndex = 0

	w.Style = Style{
		FocusedWidget:  lipgloss.NewStyle().Border(lipgloss.NormalBorder()),
		BlurredWidget:  lipgloss.NewStyle().Border(lipgloss.HiddenBorder()),
		BlurredControl: lipgloss.NewStyle().Italic(true),
		BlurredValue:   lipgloss.NewStyle().Italic(true),
		FocusedControl: lipgloss.NewStyle().Bold(true),
		FocusedValue:   lipgloss.NewStyle().Bold(true),
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

func (w *Widget[T]) Init() tea.Cmd {
	return nil
}

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

func (w *Widget[T]) Render(size orvyn.Size) string {
	var b strings.Builder
	var margin int

	margin += w.widgetStyle.GetBorderLeftSize()
	margin += w.widgetStyle.GetBorderRightSize()
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

func (w *Widget[T]) Resize(size orvyn.Size) {}

func (w *Widget[T]) GetSize() orvyn.Size {
	return orvyn.NewSize(0, 0)
}

func (w *Widget[T]) GetMinSize() orvyn.Size {
	return orvyn.NewSize(0, 0)
}

func (w *Widget[T]) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(0, 0)
}

func (w *Widget[T]) GetMaxSize() orvyn.Size {
	return orvyn.NewSize(0, 0)
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
