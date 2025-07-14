package mailwriter

import (
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/widgetfocusmanager"
	"farental/model"
	"farental/model/widget/textarea"
	"farental/model/widget/textinput"
	"farental/style"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
)

type Model struct {
	widgetfocusmanager.BaseFocusWidget

	TIReceiver *textinput.Model
	TISubject  *textinput.Model
	TIContent  *textarea.Model

	focusManager *widgetfocusmanager.WidgetFocusManager
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
	normalModeKeymap.NewKeyBinding(keybind.Help, true)

	bubblehelp.RegisterContext(model.ContextMailWriterNormalMode, normalModeKeymap)

	editModeKeymap := bubblehelp.NewKeymap(2)
	editModeKeymap.Style = style.MainHelpStyle
	editModeKeymap.NewKeyBinding(keybind.Tab, true)
	editModeKeymap.NewKeyBinding(keybind.ShiftTab, false)
	editModeKeymap.NewKeyBinding(keybind.Esc, true)
	editModeKeymap.SetHelpDesc(keybind.Esc, lang.L("stop editing"))
	editModeKeymap.NewKeyBinding(keybind.Quit, false)
	editModeKeymap.NewKeyBinding(keybind.Help, true)

	bubblehelp.RegisterContext(model.ContextMailWriterEditMode, editModeKeymap)

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

	m.focusManager = widgetfocusmanager.New()

	m.focusManager.Add(m.TIReceiver)
	m.focusManager.Add(m.TISubject)
	m.focusManager.Add(m.TIContent)

	return m
}

func (m Model) Init() tea.Cmd {
	return model.InitCmd
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case model.InitMsg:
		m.EditMode = false
		m.focusManager.BlurCurrent()
		bubblehelp.SwitchContext(model.ContextMailWriterNormalMode)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return m, tea.Quit
		case key.Matches(msg, keybind.Help):
			bubblehelp.ShowAll = !bubblehelp.ShowAll
			return m, nil
		}
	}

	if m.EditMode {
		return m.editModeUpdate(msg)
	}

	return m.normalModeUpdate(msg)
}

func (m *Model) editModeUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Esc):
			m.ExitEditMode()

			return m, nil
		}
	}

	m.focusManager.Update(msg)

	return m, nil
}

func (m *Model) normalModeUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Esc):
			return m, model.BackCmd
		}
	}

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
	bubblehelp.SwitchContext(model.ContextMailWriterNormalMode)
}

func (m *Model) Blur() {
	m.BaseFocusWidget.Blur()
}

func (m *Model) GetEditModeKeybind() *key.Binding {
	return &keybind.EKey
}

func (m *Model) EnterEditMode() {
	m.BaseFocusWidget.EnterEditMode()
	m.focusManager.Focus(0)
	bubblehelp.SwitchContext(model.ContextMailWriterEditMode)
}

func (m *Model) ExitEditMode() {
	m.BaseFocusWidget.ExitEditMode()
	m.focusManager.BlurCurrent()
	bubblehelp.SwitchContext(model.ContextMailWriterNormalMode)
}
