package textinput

import (
	"farental/internal/orvyn"
	"farental/style"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Widget struct {
	orvyn.BaseFocusable
	textinput.Model
}

func New() *Widget {
	m := new(Widget)

	m.Model = textinput.New()
	style.SetTextInputStyle(&m.Model)

	return m
}

func (m *Widget) Init() tea.Cmd {
	m.Model.SetValue("")
	return textinput.Blink
}

func (m *Widget) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	m.Model, cmd = m.Model.Update(msg)

	return cmd
}

func (m *Widget) OnFocus() {
	m.Model.Focus()
}

func (m *Widget) OnBlur() {
	m.Model.Blur()
}

func (m *Widget) Render(size orvyn.Size) string {
	var border lipgloss.Style

	if m.IsFocused() {
		border = style.FocusedStyle
	} else {
		border = style.BlurredStyle
	}

	return border.Render(m.Model.View())
}

func (m *Widget) Resize(size orvyn.Size) {
	// Take borders into account
	m.Model.Width = size.Width - 1

	// For the Bubbles textinput to process the update
	focused := m.Model.Focused()
	if !focused {
		m.Model.Focus()
	}

	m.Model, _ = m.Model.Update(nil)

	if !focused {
		m.Model.Blur()
	}
}

func (m *Widget) GetSize() orvyn.Size {
	// Take borders into account
	return orvyn.NewSize(m.Model.Width+1, 3)
}

func (m *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(26, 3)
}

func (m *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(46, 3)
}

func (m *Widget) GetMaxSize() orvyn.Size {
	return orvyn.NewSize(95, 3)
}

func (m *Widget) OnEnterInput() {}

func (m *Widget) OnExitInput() {}
