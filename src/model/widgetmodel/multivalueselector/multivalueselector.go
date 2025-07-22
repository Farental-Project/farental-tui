package multivalueselector

import (
	"farental/internal/keybind"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Value interface {
	RenderValue() string
}

type Style struct {
	BlurredControl lipgloss.Style
	FocusedControl lipgloss.Style
	BlurredValue   lipgloss.Style
	FocusedValue   lipgloss.Style
}

type Model[T Value] struct {
	Style Style

	values map[string]T
	keys   []string

	selectedIndex int

	focus bool

	Width int

	InfiniteLoop bool
}

func New[T Value]() Model[T] {
	m := Model[T]{}

	m.values = make(map[string]T)
	m.keys = make([]string, 0)

	m.selectedIndex = 0

	m.Style = Style{
		BlurredControl: lipgloss.NewStyle().Italic(true),
		BlurredValue:   lipgloss.NewStyle().Italic(true),
		FocusedControl: lipgloss.NewStyle().Bold(true),
		FocusedValue:   lipgloss.NewStyle().Bold(true),
	}

	m.InfiniteLoop = false

	return m
}

func (m Model[T]) Init() tea.Cmd {
	return nil
}

func (m Model[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focus {
			return m, nil
		}

		switch {
		case key.Matches(msg, keybind.Left):
			m.selectedIndex--

			if m.selectedIndex < 0 {
				if m.InfiniteLoop {
					m.selectedIndex = len(m.keys) - 1
				} else {
					m.selectedIndex = 0
				}
			}

		case key.Matches(msg, keybind.Right):
			m.selectedIndex++

			if m.selectedIndex > len(m.keys)-1 {
				if m.InfiniteLoop {
					m.selectedIndex = 0
				} else {
					m.selectedIndex = len(m.keys) - 1
				}
			}
		}
	}

	return m, nil
}

func (m Model[T]) View() string {
	var b strings.Builder
	var v, c lipgloss.Style

	if m.focus {
		v = m.Style.FocusedValue
		c = m.Style.FocusedControl
	} else {
		v = m.Style.BlurredValue
		c = m.Style.BlurredControl
	}

	b.WriteString(c.Render("<"))
	b.WriteString(v.Width(m.Width - 2).
		AlignHorizontal(lipgloss.Center).
		Render(m.GetSelectedValue().RenderValue()))
	b.WriteString(c.Render(">"))

	return lipgloss.NewStyle().Width(m.Width).AlignHorizontal(lipgloss.Center).Render(b.String())
}

func (m *Model[T]) SetValues(keys []string, values map[string]T) {
	m.keys = keys
	m.values = values
}

func (m Model[T]) GetSelectedValue() T {
	var empty T

	if len(m.keys) == 0 {
		return empty
	}

	return m.values[m.keys[m.selectedIndex]]
}

func (m *Model[T]) SetSelectedIndex(index int) {
	if index < 0 || index >= len(m.keys) {
		return
	}

	m.selectedIndex = index
}

func (m *Model[T]) Focus() tea.Cmd {
	m.focus = true
	return nil
}

func (m *Model[T]) Blur() {
	m.focus = false
}

func (m Model[T]) Focused() bool {
	return m.focus
}
