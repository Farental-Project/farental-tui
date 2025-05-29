package simplelogviewer

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Model struct {
	Content []string

	Viewport viewport.Model
}

func New(width, heigh int) Model {
	m := Model{
		Viewport: viewport.New(width, heigh),
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
	return m.Viewport.View()
}

func (m *Model) SetContent(content []string) {
	m.Content = content
	m.refresh()
}

func (m *Model) AddContent(content string) {
	m.Content = append(m.Content, content)
	m.refresh()
}

func (m *Model) refresh() {
	m.Viewport.SetContent(lipgloss.NewStyle().
		Width(m.Viewport.Width).Render(
		strings.Join(m.Content, "\n")))
	m.Viewport.GotoBottom()
}
