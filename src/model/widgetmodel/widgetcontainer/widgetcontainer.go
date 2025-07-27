package widgetcontainer

import (
	"farental/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Model struct {
	Title   string
	Content tea.Model
	Width   int
	Height  int
}

func New(content tea.Model, title string, width, height int) Model {
	return Model{
		Title:   title,
		Content: content,
		Width:   width,
		Height:  height,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	if len(m.Title) > 0 {
		b.WriteString(style.ContainerTitleStyle.Width(m.Width).Render(m.Title))
		b.WriteString("\n")
	}

	if m.Content != nil {
		b.WriteString(m.Content.View())
	}

	return style.FocusedStyle.
		Width(m.Width).Height(m.Height).Render(b.String())
}

func (m Model) ViewContent(s string, alignH, alignV lipgloss.Position) string {
	var b strings.Builder
	var height int

	height = m.Height

	if len(m.Title) > 0 {
		b.WriteString(style.ContainerTitleStyle.Width(m.Width).Render(m.Title))
		b.WriteString("\n")
		height -= 2
	}
	b.WriteString(style.TextStyle.Width(m.Width).Height(height).
		Align(alignH, alignV).Render(s))

	return style.FocusedStyle.
		Width(m.Width).Height(m.Height).Render(b.String())
}

func (m *Model) UpdateContent(model tea.Model) {
	m.Content = model
}
