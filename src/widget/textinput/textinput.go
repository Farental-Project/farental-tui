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
	return nil
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

func (m *Widget) Render(size *orvyn.Size) string {
	var border lipgloss.Style

	if m.IsFocused() {
		border = style.FocusedStyle
	} else {
		border = style.BlurredStyle
	}

	return border.Width(size.Width).
		Height(size.Height).Render(m.Model.View())
}

func (m *Widget) GetMinSize() orvyn.Size {
	return *orvyn.NewSize(20, 1)
}

func (m *Widget) GetPreferredSize() orvyn.Size {
	return *orvyn.NewSize(20, 1)
}

func (m *Widget) GetMaxSize() orvyn.Size {
	return *orvyn.NewSize(20, 1)
}
