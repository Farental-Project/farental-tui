package textinput

import (
	"farental/internal/widgetfocusmanager"
	"farental/style"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	widgetfocusmanager.BaseFocusWidget
	textinput.Model
}

func New() *Model {
	m := &Model{}

	m.Model = textinput.New()
	style.SetTextInputStyle(&m.Model)

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.Model, cmd = m.Model.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	return m.Model.View()
}

func (m *Model) Focus() {
	m.BaseFocusWidget.Focus()
	m.Model.Focus()
}

func (m *Model) Blur() {
	m.BaseFocusWidget.Blur()
	m.Model.Blur()
}
