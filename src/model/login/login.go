package login

import (
	"errors"
	"farental/art"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/context"
	"farental/internal/lang"
	"farental/model"
	"farental/style"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	"log"
	"strings"
)

type Model struct {
	Err    error
	Inputs [2]textinput.Model
	Focus  int
	Title  string

	Help   help.Model
	Keymap config.ModularKeyMap
}

func New() Model {
	tiUserEmail := textinput.New()
	tiUserEmail.Placeholder = lang.L("E-mail")
	tiUserEmail.Focus()
	tiUserEmail.Width = 30

	tiPassword := textinput.New()
	tiPassword.Placeholder = lang.L("Password")
	tiPassword.EchoMode = textinput.EchoPassword
	tiPassword.EchoCharacter = '*'
	tiPassword.Width = 30

	title := art.CreateASCIIArtTitle("FARENTAL")

	m := Model{}
	m.Inputs[0] = tiUserEmail
	m.Inputs[1] = tiPassword
	m.Title = title
	m.Help = help.New()

	m.Keymap = config.ModularKeyMap{}

	m.Keymap.SetBindings([][]key.Binding{
		{
			config.Tab,
			config.ShiftTab,
			config.Submit,
		},
		{
			config.Help,
			config.Quit,
		},
	})
	m.Keymap.SetEssentialBindings([]key.Binding{
		config.Submit,
		config.Help,
		config.Quit,
	})

	return m
}

func (m Model) Init() tea.Cmd {
	return model.InitCmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case model.InitMsg:
		var lastUsedEmail string

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
		case key.Matches(msg, config.Quit):
			return m, tea.Quit
		case key.Matches(msg, config.Submit):
			ret := m.submit()

			if ret {
				return context.ContentManager.SwitchContent(model.ContentCharacterSelection)
			}

			return m, nil
		case key.Matches(msg, config.Tab, config.ShiftTab):

			if key.Matches(msg, config.Tab) {
				m.Focus++
			} else if key.Matches(msg, config.ShiftTab) {
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
		case key.Matches(msg, config.Help):
			m.Help.ShowAll = !m.Help.ShowAll

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

	if m.Err != nil {
		tui.WriteString("\n\n")
		tui.WriteString(style.ErrorStyle.Render(m.Err.Error()))
	}

	tui.WriteString("\n\n\n")
	tui.WriteString(m.Help.View(m.Keymap))

	return lipgloss.Place(
		context.ContentManager.ScreenWidth, context.ContentManager.ScreenHeight,
		lipgloss.Center, lipgloss.Center,
		tui.String())
}

func (m *Model) submit() bool {
	email := m.Inputs[0].Value()
	password := m.Inputs[1].Value()

	if len(email) == 0 || len(password) == 0 {
		m.Err = errors.New(lang.L("please input e-mail and password"))
		return false
	}

	req := request.Login()

	resp, err := req.SetBody(
		api.AuthLoginBody{
			Email:    email,
			Password: password,
		}).Send()

	if err != nil {
		m.Err = errors.New(lang.L("cannot connect to Farental's server"))
		return false
	}

	if resp.StatusCode() != 200 {
		m.Err = errors.New(lang.L("invalid e-mail / password combination"))
		return false
	}

	context.Client.SetCookie(resp.Cookies()[0])

	// TODO: Manage the currently selected character to go directly to the gamedashboard.
	viper.Set("lastusedemail", email)
	err = viper.WriteConfig()

	if err != nil {
		log.Println("could not save last used e-mail : ", err)
	}

	return true
}
