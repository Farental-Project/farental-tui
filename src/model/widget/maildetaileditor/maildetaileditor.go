package maildetaileditor

import (
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/widgetfocusmanager"
	"farental/model"
	"farental/model/widget/list"
	"farental/model/widget/textinput"
	"farental/style"
	"github.com/charmbracelet/bubbles/key"
	teaList "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"strings"
)

type Model struct {
	widgetfocusmanager.BaseFocusWidget

	TIMoneyAmount   *textinput.Model
	ListAttachments *list.Model
	Attachments     *[]teaList.Item

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

	bubblehelp.RegisterContext(model.ContextMailDetailEditorEditMode, editModeKeymap)

	m := &Model{
		Width: width,
	}

	m.TIMoneyAmount = textinput.New()
	m.TIMoneyAmount.Placeholder = lang.L("Money amount")
	m.TIMoneyAmount.Prompt = ""
	m.TIMoneyAmount.Width = m.Width
	m.TIMoneyAmount.Validate = model.NumericalValidate

	m.Attachments = &[]teaList.Item{
		ListItem{
			3, "Truite",
		},
		ListItem{
			5, "Fruit",
		},
		ListItem{
			3, "Toto",
		},
		ListItem{
			2, "Tissu",
		},
		ListItem{
			3, "Patate",
		},
	}

	m.ListAttachments = list.New(*m.Attachments,
		ListItemDelegate{},
		m.Width, 5)

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

	containerStyle = style.BlurContainerStyle

	if m.Focused {
		containerStyle = style.ContainerStyle
	}

	moneyAmountField.WriteString(m.TIMoneyAmount.View())

	if m.TIMoneyAmount.Err != nil {
		moneyAmountField.WriteString("\n")
		moneyAmountField.WriteString(style.ErrorStyle.Render(m.TIMoneyAmount.Err.Error()))
	}

	tui := lipgloss.JoinVertical(lipgloss.Top,
		moneyAmountField.String(),
		m.ListAttachments.View(),
	)

	return containerStyle.Render(tui)
}

func (m *Model) Focus() {
	m.BaseFocusWidget.Focus()
	bubblehelp.SwitchContext(model.ContextMailWidgetNormalMode)
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
	bubblehelp.SwitchContext(model.ContextMailDetailEditorEditMode)
}

func (m *Model) ExitEditMode() {
	m.BaseFocusWidget.ExitEditMode()
	m.focusManager.BlurCurrent()
	bubblehelp.SwitchContext(model.ContextMailWidgetNormalMode)
}
