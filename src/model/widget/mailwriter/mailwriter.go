package mailwriter

import (
	"farental/internal/lang"
	"farental/internal/widgetfocusmanager"
	"farental/model/widget/textarea"
	"farental/model/widget/textinput"
	"farental/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	widgetfocusmanager.BaseFocusWidget

	TIReceiver *textinput.Model
	TISubject  *textinput.Model
	TIContent  *textarea.Model
}

// New creates a new Mail Writer widget, Focusable Widgets needs to return as pointer.
func New() *Model {
	m := &Model{}

	m.TIReceiver = textinput.New()
	m.TIReceiver.Placeholder = lang.L("Receiver name")
	m.TIReceiver.Prompt = ""

	m.TISubject = textinput.New()
	m.TISubject.Placeholder = lang.L("Subject")
	m.TISubject.Prompt = ""

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
	var containerStyle lipgloss.Style

	containerStyle = style.BlurContainerStyle

	if m.Focused {
		containerStyle = style.ContainerStyle
	}

	tui := lipgloss.JoinVertical(lipgloss.Top,
		m.TIReceiver.View(),
		m.TISubject.View(),
		m.TIContent.View(),
	)

	return containerStyle.Render(tui)

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
