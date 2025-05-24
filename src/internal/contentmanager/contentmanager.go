package contentmanager

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Manager struct {
	Contents     map[string]tea.Model
	CurrentCode  string
	PreviousCode string
}

func New() *Manager {
	return &Manager{
		Contents: make(map[string]tea.Model),
	}
}

func (m *Manager) RegisterContent(code string, content tea.Model) {
	m.Contents[code] = content
}

func (m *Manager) SwitchContent(code string) {
	m.PreviousCode = m.CurrentCode
	m.CurrentCode = code
}

func (m *Manager) GetCurrentModel() tea.Model {
	return m.Contents[m.CurrentCode]
}

func (m *Manager) Back() {
	m.CurrentCode = m.PreviousCode
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.CurrentCode == "" {
		return nil, nil
	}

	m.Contents[m.CurrentCode], cmd = m.Contents[m.CurrentCode].Update(msg)

	return m.Contents[m.CurrentCode], cmd
}

func (m Manager) View() string {
	if m.CurrentCode == "" {
		return ""
	}

	return m.Contents[m.CurrentCode].View()
}
