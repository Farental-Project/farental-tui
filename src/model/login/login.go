package login

import (
	"errors"
	"farental/art"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/model"
	"farental/style"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"strings"
)

type Model struct {
	ErrMsg error
	Inputs [2]textinput.Model
	Focus  int
	Title  string

	Keymap config.ModularKeyMap
}

func New() Model {
	tiUserEmail := textinput.New()
	tiUserEmail.Placeholder = lang.L("E-mail")
	tiUserEmail.Focus()
	tiUserEmail.Width = 30
	tiUserEmail.Prompt = ""
	style.SetTextInputStyle(&tiUserEmail)

	tiPassword := textinput.New()
	tiPassword.Placeholder = lang.L("Password")
	tiPassword.EchoMode = textinput.EchoPassword
	tiPassword.EchoCharacter = '*'
	tiPassword.Width = 30
	tiPassword.Prompt = ""
	style.SetTextInputStyle(&tiPassword)

	title := art.CreateASCIIArtBrokenTitle("FARENTAL")

	m := Model{}
	m.Inputs[0] = tiUserEmail
	m.Inputs[1] = tiPassword
	m.Title = title

	return m
}

func (m Model) Init() tea.Cmd {
	return model.InitCmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	defer context.ContentManager.UpdateCurrentContent(m)

	switch msg := msg.(type) {
	case model.InitMsg:
		var lastUsedEmail string
		var loginToken string

		context.KeymapManager.SwitchContext(model.ContextLogin)

		context.Client.Cookies = make([]*http.Cookie, 0)

		loginToken = viper.GetString("logintoken")

		if loginToken != "" {
			ok := m.skipLogin(loginToken)

			if ok {
				return m.nextContent()
			}

			// Expired token
			viper.Set("logintoken", "")
		}

		lastUsedEmail = viper.GetString("lastusedemail")

		if lastUsedEmail == "" {
			return m, nil
		}

		m.Inputs[0].SetValue(lastUsedEmail)
		m.Inputs[0].Blur()
		m.Inputs[1].Focus()
		m.Focus = 1

		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return m, tea.Quit
		case key.Matches(msg, keybind.Enter):
			ret := m.submit()

			if ret {
				return m.nextContent()
			}

			return m, nil
		case key.Matches(msg, keybind.Tab, keybind.ShiftTab):

			if key.Matches(msg, keybind.Tab) {
				m.Focus++
			} else if key.Matches(msg, keybind.ShiftTab) {
				m.Focus--
			}

			if m.Focus > len(m.Inputs)-1 {
				m.Focus = 0
			} else if m.Focus < 0 {
				m.Focus = len(m.Inputs) - 1
			}

			var cmd tea.Cmd

			for i := 0; i < len(m.Inputs); i++ {
				if i == m.Focus {
					cmd = m.Inputs[i].Focus()
					continue
				}

				m.Inputs[i].Blur()
			}

			return m, cmd
		case key.Matches(msg, keybind.Help):
			context.KeymapManager.ShowAll = !context.KeymapManager.ShowAll

			return m, nil
		}

	}

	context.ContentManager.Update(msg)

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.Inputs))

	for i := range m.Inputs {
		m.Inputs[i], cmds[i] = m.Inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m Model) View() string {
	var form strings.Builder
	var tui strings.Builder

	for i, input := range m.Inputs {
		var s lipgloss.Style

		if i == m.Focus {
			s = style.FocusedStyle
		} else {
			s = style.BlurredStyle
		}

		field := s.Render(input.View())
		form.WriteString(field)

		if i < len(m.Inputs)-1 {
			form.WriteString("\n")
		}
	}

	tui.WriteString(style.TitleStyle.Render(m.Title))
	tui.WriteString("\n\n\n")
	tui.WriteString(form.String())

	if m.ErrMsg != nil {
		tui.WriteString("\n\n")
		tui.WriteString(style.ErrorStyle.Render(m.ErrMsg.Error()))
	}

	tui.WriteString("\n\n\n")
	tui.WriteString(context.KeymapManager.View(style.LayoutWidth))

	return lipgloss.Place(
		context.ContentManager.ScreenWidth, context.ContentManager.ScreenHeight,
		lipgloss.Center, lipgloss.Center,
		tui.String())
}

func (m *Model) skipLogin(token string) bool {
	cookie := http.Cookie{
		Name:   "jwt",
		Value:  token,
		Secure: true,
	}

	context.Client.SetCookie(&cookie)

	ok := m.getActiveCharacter()

	if !ok {
		// Clear the bad cookie
		context.Client.Cookies = make([]*http.Cookie, 0)
	}

	return ok
}

func (m *Model) submit() bool {
	email := m.Inputs[0].Value()
	password := m.Inputs[1].Value()

	if len(email) == 0 || len(password) == 0 {
		m.ErrMsg = errors.New(lang.L("please input e-mail and password"))
		return false
	}

	req := request.Login()

	resp, err := req.SetBody(
		api.AuthLoginBody{
			Email:    email,
			Password: password,
		}).Send()

	if err != nil {
		m.ErrMsg = helper.ConnectionError()
		return false
	}

	m.ErrMsg = helper.ExtractError(resp)

	if m.ErrMsg != nil {
		return false
	}

	data := resp.Result().(*api.AuthSuccessResponse)

	viper.Set("logintoken", data.Data)

	context.Client.SetCookie(resp.Cookies()[0])

	viper.Set("lastusedemail", email)
	err = viper.WriteConfig()

	if err != nil {
		log.Println(lang.L("could not save config : "), err)
	}

	m.getActiveCharacter()

	m.Inputs[1].SetValue("")

	return true
}

func (m *Model) getActiveCharacter() bool {
	resp, err := helper.SendRequest(request.CharacterGetActive())

	if err != nil {
		m.ErrMsg = err
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

func (m *Model) nextContent() (tea.Model, tea.Cmd) {
	if context.CharacterID == 0 {
		return context.ContentManager.SwitchContent(m, model.ContentCharacterSelection)
	} else {
		return context.ContentManager.SwitchContent(m, model.ContentGameDashboard)
	}
}
