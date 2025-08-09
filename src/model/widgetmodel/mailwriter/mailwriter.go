package mailwriter

import (
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/widgetfocusmanager"
	"farental/model"
	"farental/model/widgetmodel/textarea"
	"farental/model/widgetmodel/textinput"
	"farental/style"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
)

type Model struct {
	widgetfocusmanager.BaseFocusableWidget

	TIReceiver *textinput.Model
	TISubject  *textinput.Model
	TIContent  *textarea.Model

	focusManager *widgetfocusmanager.WidgetFocusManager

	width int
}

// New creates a new Mail Writer widget, Focusable Widgets needs to return as pointer.
func New(width int) *Model {
	editModeKeymap := bubblehelp.NewKeymap(2)
	editModeKeymap.Style = style.MainHelpStyle
	editModeKeymap.NewKeyBinding(keybind.Tab, true)
	editModeKeymap.NewKeyBinding(keybind.ShiftTab, false)
	editModeKeymap.NewKeyBinding(keybind.Esc, true)
	editModeKeymap.SetHelpDesc(keybind.Esc, lang.L("stop editing"))
	editModeKeymap.NewKeyBinding(keybind.Quit, false)
	editModeKeymap.NewKeyBinding(keybind.Help, true)

	bubblehelp.RegisterContext(keybind.ContextMailWriterEditMode, editModeKeymap)

	m := &Model{width: width}

	m.TIReceiver = textinput.New()
	m.TIReceiver.Placeholder = lang.L("Receiver name")
	m.TIReceiver.Prompt = ""
	m.TIReceiver.Width = width

	m.TISubject = textinput.New()
	m.TISubject.Placeholder = lang.L("Subject")
	m.TISubject.Prompt = ""
	m.TISubject.Width = width

	m.TIContent = textarea.New()
	m.TIContent.ShowLineNumbers = false
	m.TIContent.Prompt = ""
	m.TIContent.Placeholder = lang.L("Mail content")
	m.TIContent.SetWidth(width + 1)
	m.TIContent.SetHeight(20)

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
		bubblehelp.SwitchContext(keybind.ContextMailWriterNormalMode)
		m.EditMode = false
		m.focusManager.BlurCurrent()

		m.TIReceiver.SetValue("")
		m.TISubject.SetValue("")
		m.TIContent.SetValue("")

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return m, tea.Quit
		case key.Matches(msg, keybind.Help):
			if !m.TIContent.Model.Focused() {
				bubblehelp.ShowAll = !bubblehelp.ShowAll
				return m, nil
			}
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

	containerStyle = style.BlurredStyle

	if m.Focused {
		containerStyle = style.FocusedStyle
	}

	tui := lipgloss.JoinVertical(lipgloss.Top,
		m.TIReceiver.View(),
		m.TISubject.View(),
		m.TIContent.View(),
	)

	return containerStyle.Render(tui)

}

func (m *Model) Focus() {
	m.BaseFocusableWidget.Focus()
	bubblehelp.SwitchContext(keybind.ContextMailWriterNormalMode)
}

func (m *Model) Blur() {
	m.BaseFocusableWidget.Blur()
}

func (m *Model) GetEditModeKeybind() *key.Binding {
	return &keybind.EKey
}

func (m *Model) EnterEditMode() {
	m.BaseFocusableWidget.EnterEditMode()
	m.focusManager.Focus(0)
	bubblehelp.SwitchContext(keybind.ContextMailWriterEditMode)
}

func (m *Model) ExitEditMode() {
	m.BaseFocusableWidget.ExitEditMode()
	m.focusManager.BlurCurrent()
	bubblehelp.SwitchContext(keybind.ContextMailWriterNormalMode)
}
