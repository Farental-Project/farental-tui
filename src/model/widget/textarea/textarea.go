package textarea

import (
	"farental/internal/widgetfocusmanager"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	textarea.Model
	widgetfocusmanager.BaseFocusableWidget
}

func New() *Model {
	m := &Model{}

	m.Model = textarea.New()

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.Model, cmd = m.Model.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	return m.Model.View()
}

func (m *Model) Focus() {
	m.BaseFocusableWidget.Focus()
	m.Model.Focus()
}

func (m *Model) Blur() {
	m.BaseFocusableWidget.Blur()
	m.Model.Blur()
}
