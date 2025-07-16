package maileditor

import (
	"farental/internal/context"
	"farental/internal/widgetfocusmanager"
	"farental/model"
	"farental/model/widget/maildetaileditor"
	"farental/model/widget/mailwriter"
	"farental/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
)

type Model struct {
	focusManager *widgetfocusmanager.WidgetFocusManager

	MailWriter       *mailwriter.Model
	MailDetailEditor *maildetaileditor.Model
}

func New() Model {
	m := Model{
		focusManager:     widgetfocusmanager.New(),
		MailWriter:       mailwriter.New(),
		MailDetailEditor: maildetaileditor.New(),
	}

	m.focusManager.Add(m.MailWriter)
	m.focusManager.Add(m.MailDetailEditor)

	m.focusManager.Focus(0)

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.MailWriter.Init(), m.MailDetailEditor.Init())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case model.BackMsg:
		return context.ContentManager.Back(m)
	}

	cmd := m.focusManager.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		m.MailWriter.View(),
		m.MailDetailEditor.View(),
		bubblehelp.View(style.LayoutWidth),
	)
}
