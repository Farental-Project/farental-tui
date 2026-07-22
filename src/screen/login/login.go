package login

import (
	"farental/art"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/session"
	"farental/screen"
	"farental/widget"
	"farental/widget/help"
	"farental/widget/languageindicator"
	"fmt"
	"log"
	"net/http"

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
	"github.com/spf13/viper"
)

type gotoCharacterSelectionMsg int

func gotoCharacterSelectionCmd() tea.Msg {
	return gotoCharacterSelectionMsg(1)
}

type Screen struct {
	title   *orvyn.SimpleRenderable
	version *orvyn.SimpleRenderable
	server  *orvyn.SimpleRenderable

	tiEmail    *textinput.Widget
	tiPassword *textinput.Widget

	languageIndicator *languageindicator.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout

	focusManager *orvyn.FocusManager
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable(fmt.Sprintf("%s",
		t.Style(theme.TitleStyleID).Render(art.CreateASCIIArtBrokenTitle("farental"))))

	s.version = orvyn.NewSimpleRenderable(config.VERSION)
	s.server = orvyn.NewSimpleRenderable(config.BaseURL)
	s.server.SetActive(false)
	s.version.Style = t.Style(theme.DimTextStyleID)

	s.tiEmail = textinput.New()

	s.tiPassword = textinput.New()
	s.tiPassword.EchoMode = teatextinput.EchoPassword
	s.tiPassword.EchoCharacter = art.CharBullet

	s.languageIndicator = languageindicator.New()

	s.statusMessage = statusmessage.New()
	s.statusMessage.SetMinSize(orvyn.NewSize(30, 1))

	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewVBoxLayout(10,
			s.title,
			orvyn.VGap,
			s.version,
			s.server,
			orvyn.VGap,
			s.tiEmail,
			s.tiPassword,
			orvyn.VGap,
			s.languageIndicator,
			orvyn.VGap,
			s.statusMessage,
			s.help,
		),
	)

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.tiEmail)
	s.focusManager.Add(s.tiPassword)
	s.focusManager.Focus(0)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	var cmds []tea.Cmd

	s.statusMessage.Reset()

	switch param := i.(type) {
	case error:
		s.statusMessage.SetError(param)
	case widget.StatusMessageParam:
		s.statusMessage.SetMessage(param.Content, param.Type)
	}

	if session.TakeExpired() {
		s.statusMessage.SetMessage(
			lokyn.L(session.ExpiredMessage),
			statusmessage.WarningMessage)
	}

	s.tiEmail.Placeholder = lokyn.L("Email")
	s.tiPassword.Placeholder = lokyn.L("Password")

	cmds = append(cmds, s.tiEmail.Init())
	cmds = append(cmds, s.tiPassword.Init())

	bubblehelp.SwitchContext(keybind.ContextLogin)

	context.Client.Cookies = make([]*http.Cookie, 0)

	loginToken := viper.GetString("logintoken")

	if loginToken != "" {
		ok := s.skipLogin(loginToken)

		if ok {
			return gotoCharacterSelectionCmd
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

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.LKeyCtrl):
			s.languageIndicator.SwitchLanguage()
			return orvyn.SwitchScreen(screen.IDLogin)

		case key.Matches(msg, keybind.NKeyCtrl):
			return orvyn.SwitchScreen(screen.IDAccountCreation)

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

	case gotoCharacterSelectionMsg:
		return s.nextScreen()
	}

	s.focusManager.Update(msg)

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
	// Set the right language
	info, err := helper.Fetch[api.UserResponse](request.AuthInfo())

	if err != nil {
		return orvyn.SwitchScreen(screen.IDCharacterSelection)
	}

	config.ChangeLanguage(info.LanguageCode)

	return orvyn.SwitchScreen(screen.IDCharacterSelection)
}
