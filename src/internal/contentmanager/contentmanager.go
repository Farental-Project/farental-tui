package contentmanager

import (
	tea "github.com/charmbracelet/bubbletea"
	"log"
)

type Manager struct {
	Contents     map[string]tea.Model
	CurrentCode  string
	PreviousCode string

	ScreenWidth, ScreenHeight int
}

func New() *Manager {
	return &Manager{
		Contents: make(map[string]tea.Model),
	}
}

func (m *Manager) RegisterContent(code string, content tea.Model) {
	m.Contents[code] = content
}

func (m *Manager) SwitchContent(code string) (tea.Model, tea.Cmd) {
	m.PreviousCode = m.CurrentCode
	m.CurrentCode = code

	_, ok := m.Contents[code]

	if !ok {
		log.Fatal("No such content : ", code)
		return nil, nil
	}

	cmd := m.Contents[code].Init()

	return m.Contents[m.CurrentCode], cmd
}

func (m *Manager) GetCurrentModel() tea.Model {
	return m.Contents[m.CurrentCode]
}

func (m *Manager) Back() (tea.Model, tea.Cmd) {
	return m.SwitchContent(m.PreviousCode)
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ScreenWidth = msg.Width
		m.ScreenHeight = msg.Height
		return m.Contents[m.CurrentCode], nil
	}

	return m.Contents[m.CurrentCode], nil
}

func (m Manager) View() string {
	if m.CurrentCode == "" {
		return ""
	}

	return m.Contents[m.CurrentCode].View()
}
