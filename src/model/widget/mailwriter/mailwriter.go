package mailwriter

import (
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/widgetfocusmanager"
	"farental/model/widget/textarea"
	"farental/model/widget/textinput"
	"farental/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
)

type Model struct {
	widgetfocusmanager.BaseFocusWidget

	TIReceiver *textinput.Model
	TISubject  *textinput.Model
	TIContent  *textarea.Model

	editMode bool
}

// New creates a new Mail Writer widget, Focusable Widgets needs to return as pointer.
func New() *Model {
	normalModeKeymap := bubblehelp.NewKeymap(2)
	normalModeKeymap.Style = style.MainHelpStyle
	normalModeKeymap.NewKeyBinding(keybind.EKey, true)
	normalModeKeymap.SetHelpDesc(keybind.EKey, lang.L("edit"))
	normalModeKeymap.NewKeyBinding(keybind.Tab, false)
	normalModeKeymap.NewKeyBinding(keybind.ShiftTab, false)
	normalModeKeymap.NewKeyBinding(keybind.Esc, true)
	normalModeKeymap.NewKeyBinding(keybind.Quit, false)

	editModeKeymap := bubblehelp.NewKeymap(2)
	editModeKeymap.Style = style.MainHelpStyle
	editModeKeymap.NewKeyBinding(keybind.Tab, true)
	editModeKeymap.NewKeyBinding(keybind.ShiftTab, false)
	editModeKeymap.NewKeyBinding(keybind.Esc, true)
	editModeKeymap.SetHelpDesc(keybind.Esc, lang.L("stop editing"))

	m := &Model{}

	m.TIReceiver = textinput.New()
	m.TIReceiver.Placeholder = lang.L("Receiver name")
	m.TIReceiver.Prompt = ""
	m.TIReceiver.Width = 100

	m.TISubject = textinput.New()
	m.TISubject.Placeholder = lang.L("Subject")
	m.TISubject.Prompt = ""
	m.TISubject.Width = 100

	m.TIContent = textarea.New()

	m.editMode = false

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.editMode {
		return m.editModeUpdate(msg)
	}

	return m, nil
}

func (m Model) editModeUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, n
}

func (m Model) normalModeUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, n
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
