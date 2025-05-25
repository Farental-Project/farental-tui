package login

import (
	"errors"
	"farental/art"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal"
	"farental/internal/lang"
	"farental/model"
	"farental/style"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Model struct {
	Err    error
	Inputs [2]textinput.Model
	Focus  int
	Title  string

	width, height int

	ctx *internal.AppCtx
}

func New(ctx *internal.AppCtx) Model {
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

	l := Model{ctx: ctx}
	l.Inputs[0] = tiUserEmail
	l.Inputs[1] = tiPassword
	l.Title = title

	return l
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			var model tea.Model

			model = m

			ret := m.submit()

			if ret {
				model = m.ctx.ContentManager.GetCurrentModel()
			}

			return model, nil
		case "tab", "shift+tab":
			key := msg.String()

			if key == "tab" {
				m.Focus++
			} else if key == "shift+tab" {
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
		}

	}

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

	tui := style.TitleStyle.Render(m.Title) +
		"\n\n" + style.ContainerStyle.Render(form.String())

	if m.Err != nil {
		tui += "\n\n" + style.ErrorStyle.Render(m.Err.Error())
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		tui)
}

func (m *Model) submit() bool {
	email := m.Inputs[0].Value()
	password := m.Inputs[1].Value()

	if len(email) == 0 || len(password) == 0 {
		m.Err = errors.New("please input e-mail and password")
		return false
	}

	req := request.Login()

	resp, err := req.SetBody(
		api.AuthLoginBody{
			Email:    email,
			Password: password,
		}).Send()

	if err != nil {
		m.Err = err
		return false
	}

	if resp.StatusCode() != 200 {
		m.Err = errors.New("invalid e-mail / password combination")
		return false
	}

	m.ctx.Client.SetCookie(resp.Cookies()[0])

	m.ctx.ContentManager.SwitchContent(model.ContentCharacterSelection)

	return true
}
