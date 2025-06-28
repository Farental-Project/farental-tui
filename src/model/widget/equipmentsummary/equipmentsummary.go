package equipmentsummary

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	width int
}

func New(width int) Model {
	m := Model{
		width: width,
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	return "Equipments"
}
