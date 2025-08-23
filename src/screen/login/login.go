package login

import (
	"farental/art"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	layout "farental/layout"
	"farental/screen"
	"farental/style"
	"farental/widget/help"
	"farental/widget/statusmessage"
	"farental/widget/textinput"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	teatextinput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	tiEmail    *textinput.Widget
	tiPassword *textinput.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout

	focusManager *orvyn.FocusManager
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewSimpleRenderable(fmt.Sprintf("%s",
		style.TitleStyle.Render(art.CreateASCIIArtBrokenTitle("farental"))))

	s.tiEmail = textinput.New()
	s.tiEmail.Placeholder = lokyn.L("Email")

	s.tiPassword = textinput.New()
	s.tiPassword.Placeholder = lokyn.L("Password")
	s.tiPassword.EchoMode = teatextinput.EchoPassword
	s.tiPassword.EchoCharacter = art.CharBullet

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewVBoxLayout(10,
			[]orvyn.Renderable{
				s.title,
				orvyn.VGap,
				orvyn.VGap,
				s.tiEmail,
				s.tiPassword,
				orvyn.VGap,
				s.statusMessage,
				s.help,
			},
		),
	)

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.tiEmail)
	s.focusManager.Add(s.tiPassword)
	s.focusManager.Focus(0)

	return s
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit

		case key.Matches(msg, keybind.Help):
			bubblehelp.ShowAll = !bubblehelp.ShowAll

			return nil

		case key.Matches(msg, keybind.Enter):
			if s.submit() {
				return s.nextScreen()
			}

			return nil
		}
	}

	s.focusManager.Update(msg)

	return nil
}

func (s *Screen) OnEnter(_ any) tea.Cmd {
	var cmds []tea.Cmd

	cmds = append(cmds, s.tiEmail.Init())
	cmds = append(cmds, s.tiPassword.Init())

	bubblehelp.SwitchContext(keybind.ContextLogin)

	context.Client.Cookies = make([]*http.Cookie, 0)

	loginToken := viper.GetString("logintoken")

	if loginToken != "" {
		ok := s.skipLogin(loginToken)

		if ok {
			return s.nextScreen()
		}

		// Expired token
		viper.Set("logintoken", "")
	}

	lastUsedEmail := viper.GetString("lastusedemail")

	if lastUsedEmail != "" {
		s.tiEmail.SetValue(lastUsedEmail)
		s.focusManager.Focus(1)
	} else {
		s.focusManager.Focus(0)
	}

	return tea.Batch(cmds...)
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) submit() bool {
	email := s.tiEmail.Value()
	password := s.tiPassword.Value()

	if len(email) == 0 || len(password) == 0 {
		s.statusMessage.SetMessage(
			lokyn.L("Please input e-mail and password"),
			statusmessage.ErrorMessage)
		return false
	}

	req := request.Login()

	req.SetBody(
		api.AuthLoginBody{
			Email:    email,
			Password: password,
		})

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	data := resp.Result().(*api.AuthSuccessResponse)

	viper.Set("logintoken", data.Data)

	context.Client.SetCookie(resp.Cookies()[0])

	viper.Set("lastusedemail", email)
	err = viper.WriteConfig()

	if err != nil {
		log.Println(lokyn.L("could not save config : "), err)
	}

	s.getActiveCharacter()

	// Avoid keeping the password in the RAM
	s.tiPassword.SetValue("")

	return true
}

func (s *Screen) skipLogin(token string) bool {
	cookie := http.Cookie{
		Name:   "jwt",
		Value:  token,
		Secure: true,
	}

	context.Client.SetCookie(&cookie)

	ok := s.getActiveCharacter()

	if !ok {
		// Clear the bad cookie
		context.Client.Cookies = make([]*http.Cookie, 0)
	}

	return ok
}

func (s *Screen) getActiveCharacter() bool {
	resp, err := helper.SendRequest(request.CharacterGetActive())

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	context.CharacterID = 0

	if resp.StatusCode() == 200 {
		character, ok := resp.Result().(*api.CharacterBasicResponse)

		if ok {
			context.CharacterID = character.ID
			return true
		}
	}

	return false
}

func (s *Screen) nextScreen() tea.Cmd {
	return orvyn.SwitchScreen(screen.IDCharacterSelection)
}
