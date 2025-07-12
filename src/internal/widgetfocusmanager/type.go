package widgetfocusmanager

import (
	tea "github.com/charmbracelet/bubbletea"
)

type FocusableWidget interface {
	tea.Model
	Focus()
	Blur()
}

type BaseFocusManager struct {
	Focused bool
}

func (b BaseFocusManager) Init() tea.Cmd {
	return nil
}

func (b BaseFocusManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return b, nil
}

func (b BaseFocusManager) View() string {
	return ""
}

func (b *BaseFocusManager) Focus() {
	b.Focused = true
}

func (b *BaseFocusManager) Blur() {
	b.Focused = false
}
