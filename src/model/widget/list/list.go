package list

import (
	"farental/internal/widgetfocusmanager"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	widgetfocusmanager.BaseFocusWidget
	list.Model
}

func New(items []list.Item, delegate list.ItemDelegate,
	width int, height int) *Model {
	m := &Model{}

	m.Model = list.New(
		items, delegate,
		width, height,
	)

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
	m.BaseFocusWidget.Focus()
}

func (m *Model) Blur() {
	m.BaseFocusWidget.Blur()
}
