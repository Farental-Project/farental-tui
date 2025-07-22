package maileditor

import (
	"farental/internal/context"
	"farental/internal/lang"
	"farental/internal/widgetfocusmanager"
	"farental/model"
	"farental/model/widgetmodel/maildetaileditor"
	"farental/model/widgetmodel/mailwriter"
	"farental/style"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
)

type Model struct {
	focusManager *widgetfocusmanager.WidgetFocusManager

	MailWriter       *mailwriter.Model
	MailDetailEditor *maildetaileditor.Model

	Title string
}

func New() Model {
	m := Model{
		focusManager:     widgetfocusmanager.New(),
		MailWriter:       mailwriter.New(48),
		MailDetailEditor: maildetaileditor.New(25),
		Title:            style.TitleStyle.Render(lang.L("New mail")),
	}

	m.focusManager.Add(m.MailWriter)
	m.focusManager.Add(m.MailDetailEditor)

	m.focusManager.Focus(0)

	return m
}

func (m Model) Init() tea.Cmd {
	return model.InitCmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case model.InitMsg:
		m.focusManager.Focus(0)

		m.MailWriter.Update(msg)
		m.MailDetailEditor.Update(msg)

		return m, nil

	case model.BackMsg:
		return context.ContentManager.Back(m)
	}

	cmd := m.focusManager.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(m.Title)

	b.WriteString("\n\n")

	b.WriteString(lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.MailWriter.View(),
		m.MailDetailEditor.View()))

	b.WriteString("\n\n")
	b.WriteString(bubblehelp.View(style.LayoutWidth))

	return lipgloss.Place(
		context.ContentManager.ScreenWidth, context.ContentManager.ScreenHeight,
		lipgloss.Center, lipgloss.Center, b.String())
}
