package accountcreation

import (
	"farental/art"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/widget/help"

	"github.com/charmbracelet/bubbles/key"
	teatextinput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/textinput"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	tiUsername        *textinput.Widget
	tiEmail           *textinput.Widget
	tiPassword        *textinput.Widget
	tiConfirmPassword *textinput.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout

	focusManager *orvyn.FocusManager
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable(
		t.Style(theme.TitleStyleID).Render(lokyn.L("New account")),
	)

	s.tiUsername = textinput.New()
	s.tiUsername.Placeholder = lokyn.L("Username")

	s.tiEmail = textinput.New()
	s.tiEmail.Placeholder = lokyn.L("Email address")

	s.tiPassword = textinput.New()
	s.tiPassword.Placeholder = lokyn.L("Password")
	s.tiPassword.EchoMode = teatextinput.EchoPassword
	s.tiPassword.EchoCharacter = art.CharBullet

	s.tiConfirmPassword = textinput.New()
	s.tiConfirmPassword.Placeholder = lokyn.L("Confirm password")
	s.tiConfirmPassword.EchoMode = teatextinput.EchoPassword
	s.tiConfirmPassword.EchoCharacter = art.CharBullet

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewVBoxLayout(10,
			[]orvyn.Renderable{
				s.title,
				orvyn.VGap,
				s.tiUsername,
				s.tiEmail,
				s.tiPassword,
				s.tiConfirmPassword,
				orvyn.VGap,
				s.statusMessage,
				s.help,
			},
		),
	)

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.tiUsername)
	s.focusManager.Add(s.tiEmail)
	s.focusManager.Add(s.tiPassword)
	s.focusManager.Add(s.tiConfirmPassword)
	s.focusManager.FocusFirst()

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextCharacterCreation)

	s.tiUsername.SetValue("")
	s.tiEmail.SetValue("")
	s.tiPassword.SetValue("")
	s.tiConfirmPassword.SetValue("")

	s.statusMessage.Reset()

	s.focusManager.FocusFirst()

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit

		case key.Matches(msg, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()

		case key.Matches(msg, keybind.Enter):
			ok := s.submit()

			if ok {
				return orvyn.SwitchToPreviousScreen()
			}

			return nil
		}
	}

	s.focusManager.Update(msg)

	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) submit() bool {
	// Basic validation
	username := s.tiUsername.Value()
	email := s.tiEmail.Value()
	password := s.tiPassword.Value()
	confirmPassword := s.tiConfirmPassword.Value()

	if username == "" {
		s.statusMessage.SetMessage(lokyn.L("Please enter an username"), statusmessage.ErrorMessage)
		return false
	}

	if email == "" || !helper.EmailIsValid(email) {
		s.statusMessage.SetMessage(lokyn.L("Please enter a valid email"), statusmessage.ErrorMessage)
		return false
	}

	if password == "" {
		s.statusMessage.SetMessage(lokyn.L("Please enter a valid password"), statusmessage.ErrorMessage)
		return false
	}

	if confirmPassword != password {
		s.statusMessage.SetMessage(lokyn.L("Passwords are not the same"), statusmessage.ErrorMessage)
		return false
	}

	s.statusMessage.SetMessage(lokyn.L("Data are valid, sending request..."), statusmessage.InformationMessage)

	req := request.SignUp(
		s.tiUsername.Value(),
		s.tiEmail.Value(),
		s.tiPassword.Value(),
		s.tiConfirmPassword.Value(),
	)

	_, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	return true
}
