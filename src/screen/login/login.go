package login

import (
	"farental/internal/keybind"
	"farental/internal/orvyn"
	"farental/internal/orvyn/layout/vbox"
	"farental/widget/textinput"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const ID orvyn.ScreenID = "login"

type Screen struct {
	orvyn.BaseScreen

	input1 *textinput.Widget
	input2 *textinput.Widget

	layout *vbox.Layout
}

func New() *Screen {
	s := new(Screen)

	s.input1 = textinput.New()
	s.input2 = textinput.New()

	s.layout = vbox.New([]orvyn.Renderable{s.input1, s.input2})

	return s
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	s.BaseScreen.Update(msg)

	s.layout.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit
		}
	}

	s.input1.Update(msg)
	s.input2.Update(msg)

	return nil
}

func (s *Screen) OnEnter(_ interface{}) tea.Cmd {
	return s.BaseScreen.OnEnter(s)
}

func (s *Screen) OnExit() interface{} {
	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}
