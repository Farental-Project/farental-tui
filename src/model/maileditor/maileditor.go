package maileditor

import (
	"farental/internal/keybind"
	"farental/internal/widgetfocusmanager"
	"farental/model/widget/mailwriter"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	focusManager *widgetfocusmanager.WidgetFocusManager

	MailWriter  *mailwriter.Model
	MailWriter2 *mailwriter.Model
}

func New() Model {
	m := Model{
		focusManager: widgetfocusmanager.New(),
		MailWriter:   mailwriter.New(),
		MailWriter2:  mailwriter.New(),
	}

	m.focusManager.Add(m.MailWriter)
	m.focusManager.Add(m.MailWriter2)

	m.focusManager.Focus(0)

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return m, tea.Quit

		}
	}

	m.focusManager.Update(msg)

	return m, nil
}

func (m Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		m.MailWriter.View(),
		m.MailWriter2.View(),
	)
}
