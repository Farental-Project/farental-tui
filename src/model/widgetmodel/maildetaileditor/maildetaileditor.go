package maildetaileditor

import (
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/widgetfocusmanager"
	"farental/model"
	"farental/model/widgetmodel/textinput"
	"farental/style"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"strings"
)

type Model struct {
	widgetfocusmanager.BaseFocusableWidget

	TitleAttachments string
	TIMoneyAmount    *textinput.Model
	ListAttachments  *ListAttachmentModel

	focusManager *widgetfocusmanager.WidgetFocusManager

	Width int
}

func New(width int) *Model {
	editModeKeymap := bubblehelp.NewKeymap(2)
	editModeKeymap.Style = style.MainHelpStyle
	editModeKeymap.NewKeyBinding(keybind.Tab, true)
	editModeKeymap.NewKeyBinding(keybind.ShiftTab, false)
	editModeKeymap.NewKeyBinding(keybind.Esc, true)
	editModeKeymap.SetHelpDesc(keybind.Esc, lang.L("stop editing"))
	editModeKeymap.NewKeyBinding(keybind.Quit, false)
	editModeKeymap.NewKeyBinding(keybind.Help, true)

	bubblehelp.RegisterContext(keybind.ContextMailDetailEditorEditMode, editModeKeymap)

	m := &Model{
		Width: width,
		TitleAttachments: style.DimBottomBorderStyle.Width(width).Render(
			style.DimTextStyle.Render(lang.L("Attachments"))),
	}

	m.TIMoneyAmount = textinput.New()
	m.TIMoneyAmount.Placeholder = lang.L("Money amount")
	m.TIMoneyAmount.Prompt = ""
	m.TIMoneyAmount.Width = m.Width - 3
	m.TIMoneyAmount.Validate = model.NumericalValidate

	m.ListAttachments = NewListAttachment(width-2, 5)

	m.focusManager = widgetfocusmanager.New()

	m.focusManager.Add(m.TIMoneyAmount)
	m.focusManager.Add(m.ListAttachments)

	return m
}

func (m *Model) Init() tea.Cmd {
	return model.InitCmd
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case model.InitMsg:
		m.TIMoneyAmount.SetValue("")
		m.ListAttachments.List.Blur()

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
		return m.updateEditMode(msg)
	}

	return m.updateNormalMode(msg)
}

func (m *Model) updateNormalMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Esc):
			return m, model.BackCmd
		}
	}

	return m, nil
}

func (m *Model) updateEditMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Esc):
			m.ExitEditMode()

			return m, nil
		}
	}

	cmd := m.focusManager.Update(msg)

	return m, cmd
}

func (m *Model) View() string {
	var containerStyle lipgloss.Style
	var moneyAmountField strings.Builder

	containerStyle = style.BlurredStyle

	if m.Focused {
		containerStyle = style.FocusedStyle
	}

	moneyAmountField.WriteString(m.TIMoneyAmount.View())

	if m.TIMoneyAmount.Err != nil {
		moneyAmountField.WriteString("\n")
		moneyAmountField.WriteString(style.ErrorStyle.Render(m.TIMoneyAmount.Err.Error()))
	}

	tui := lipgloss.JoinVertical(lipgloss.Top,
		m.TitleAttachments,
		moneyAmountField.String(),
		m.ListAttachments.View(),
	)

	return containerStyle.Render(tui)
}

func (m *Model) Focus() {
	m.BaseFocusableWidget.Focus()
	bubblehelp.SwitchContext(keybind.ContextMailWidgetNormalMode)
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
	bubblehelp.SwitchContext(keybind.ContextMailDetailEditorEditMode)
}

func (m *Model) ExitEditMode() {
	m.BaseFocusableWidget.ExitEditMode()
	m.focusManager.BlurCurrent()
	bubblehelp.SwitchContext(keybind.ContextMailWidgetNormalMode)
}
