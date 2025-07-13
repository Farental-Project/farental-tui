package mailwriter

import (
	"farental/internal/widgetfocusmanager"
	"farental/style"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	widgetfocusmanager.BaseFocusWidget

	TISubject textinput.Model
	TIContent textarea.Model
}

// New creates a new Mail Writer widget, Focusable Widgets needs to return as pointer.
func New() *Model {
	m := &Model{}

	m.TISubject = textinput.New()
	m.TIContent = textarea.New()

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	if m.Focused {
		return style.TitleStyle.Render("FOCUSED")
	}

	return style.DimTextStyle.Render("NOT FOCUSED")

}

func (m *Model) Focus() {
	m.BaseFocusWidget.Focus()
}

func (m *Model) Blur() {
	m.BaseFocusWidget.Blur()
}

func (m *Model) updateFocus() {
	if m.Focused {

	}
}
