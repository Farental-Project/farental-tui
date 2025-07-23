package login

import (
	"farental/art"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/internal/orvyn/layout"
	"farental/widget/textinput"
	"github.com/charmbracelet/bubbles/key"
	teatextinput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const ID orvyn.ScreenID = "login"

type Screen struct {
	orvyn.BaseScreen

	tiEmail    *textinput.Widget
	tiPassword *textinput.Widget

	layout *layout.CenterLayout

	focusManager *orvyn.FocusManager
}

func New() *Screen {
	s := new(Screen)

	s.tiEmail = textinput.New()
	s.tiEmail.Placeholder = lang.L("Email")

	s.tiPassword = textinput.New()
	s.tiPassword.Placeholder = lang.L("Password")
	s.tiPassword.EchoMode = teatextinput.EchoPassword
	s.tiPassword.EchoCharacter = art.CharBullet

	s.layout = layout.NewCenterLayout(layout.NewVBoxLayout([]orvyn.Renderable{s.tiEmail, s.tiPassword}))

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.tiEmail)
	s.focusManager.Add(s.tiPassword)
	s.focusManager.Focus(0)

	return s
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	s.BaseScreen.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit
		}
	}

	s.focusManager.Update(msg)

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
