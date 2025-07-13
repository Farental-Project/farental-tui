package widgetfocusmanager

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type FocusableWidget interface {
	tea.Model
	Focus()
	Blur()
	GetFocusKeybind() *key.Binding
}

type BaseFocusWidget struct {
	Focused bool
}

func (b BaseFocusWidget) Init() tea.Cmd {
	return nil
}

func (b BaseFocusWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return b, nil
}

func (b BaseFocusWidget) View() string {
	return ""
}

func (b *BaseFocusWidget) Focus() {
	b.Focused = true
}

func (b *BaseFocusWidget) Blur() {
	b.Focused = false
}

func (b *BaseFocusWidget) GetFocusKeybind() *key.Binding {
	return nil
}
