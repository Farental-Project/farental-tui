package simplelogviewer

import (
	"farental/style"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

var (
	styleTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(style.ColorHighlightDim)).
		Foreground(lipgloss.Color(style.ColorHighlightDim)).
		BorderTop(false).BorderRight(false).BorderLeft(false)
)

type Model struct {
	Content []string
	Title   string

	Viewport viewport.Model

	width int
}

func New(title string, width, height int) Model {
	m := Model{
		Title:    title,
		width:    width,
		Viewport: viewport.New(width, height),
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
	var b strings.Builder

	b.WriteString(styleTitle.Width(m.width).Render(m.Title))
	b.WriteString("\n")
	b.WriteString(m.Viewport.View())

	return b.String()
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
