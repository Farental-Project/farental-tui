package model

import (
	"errors"
	"farental/art"
	"farental/core/data/api"
	"farental/core/request"
	"farental/data"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Login struct {
	Err    error
	Inputs [2]textinput.Model
	Focus  int
	Title  string

	width  int
	height int

	ctx *data.AppCtx
}

func LoginModel(ctx *data.AppCtx) Login {
	tiUserEmail := textinput.New()
	tiUserEmail.Placeholder = "E-mail"
	tiUserEmail.Focus()
	tiUserEmail.Width = 30

	tiPassword := textinput.New()
	tiPassword.Placeholder = "Password"
	tiPassword.EchoMode = textinput.EchoPassword
	tiPassword.EchoCharacter = '*'
	tiPassword.Width = 30

	title := art.CreateASCIIArtTitle("FARENTAL")

	l := Login{ctx: ctx}
	l.Inputs[0] = tiUserEmail
	l.Inputs[1] = tiPassword
	l.Title = title

	return l
}

func (l Login) Init() tea.Cmd {
	return textinput.Blink
}

func (l Login) Update(msg tea.Msg) (Login, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.width = msg.Width
		l.height = msg.Height
		return l, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return l, tea.Quit
		case "enter":
			l.submit()

			return l, nil
		case "tab", "shift+tab":
			key := msg.String()

			if key == "tab" {
				l.Focus++
			} else if key == "shift+tab" {
				l.Focus--
			}

			if l.Focus > len(l.Inputs)-1 {
				l.Focus = 0
			} else if l.Focus < 0 {
				l.Focus = len(l.Inputs) - 1
			}

			var cmd tea.Cmd

			for i := 0; i < len(l.Inputs); i++ {
				if i == l.Focus {
					cmd = l.Inputs[i].Focus()
					continue
				}

				l.Inputs[i].Blur()
			}

			return l, cmd
		}

	}

	cmd := l.updateInputs(msg)
	return l, cmd
}

func (l *Login) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(l.Inputs))

	for i := range l.Inputs {
		l.Inputs[i], cmds[i] = l.Inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (l Login) View() string {
	var b strings.Builder

	for i, input := range l.Inputs {
		var style lipgloss.Style

		if i == l.Focus {
			style = focusedStyle
		} else {
			style = blurredStyle
		}

		field := style.Render(input.View())
		b.WriteString(field)

		if i < len(l.Inputs)-1 {
			b.WriteString("\n")
		}
	}

	if l.Err != nil {
		b.WriteString("\n\n" + errorStyle.Render(l.Err.Error()))
	}

	form := titleStyle.Render(l.Title) +
		"\n\n" + containerStyle.Render(b.String())

	return lipgloss.Place(
		l.width, l.height,
		lipgloss.Center, lipgloss.Center,
		form)
}

func (l *Login) submit() {
	email := l.Inputs[0].Value()
	password := l.Inputs[1].Value()

	if len(email) == 0 || len(password) == 0 {
		l.Err = errors.New("please input e-mail and password")
		return
	}

	req := request.Login()

	resp, err := req.SetBody(
		api.AuthLoginBody{
			Email:    email,
			Password: password,
		}).Send()

	if err != nil {
		l.Err = err
		return
	}

	if resp.StatusCode() != 200 {
		l.Err = errors.New("invalid e-mail / password combination")
		return
	}

	l.ctx.Client.SetCookie(resp.Cookies()[0])
	l.Err = errors.New("LOGIN SUCCESS")
}
