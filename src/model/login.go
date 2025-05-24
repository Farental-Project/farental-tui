package model

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Login struct {
	Err         error
	TIUserEmail textinput.Model
	TIPassword  textinput.Model

	width  int
	height int
}

func LoginModel() Login {
	tiUserEmail := textinput.New()
	tiUserEmail.Placeholder = "E-mail"

	tiPassword := textinput.New()
	tiPassword.Placeholder = "Password"
	tiPassword.EchoMode = textinput.EchoPassword
	tiPassword.EchoCharacter = '*'

	return Login{
		TIUserEmail: tiUserEmail,
		TIPassword:  tiPassword,
	}
}

func (l Login) Init() tea.Cmd {
	return nil
}

func (l Login) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.width = msg.Width
		l.height = msg.Height
		return l, nil
	}
}

func (l Login) View() string {
	var b strings.Builder

	title := titleStyle.Render("Login")
	b.WriteString(title + "\n\n")

	tiEmail := blurredStyle.Render(l.TIUserEmail.View())
	b.WriteString(tiEmail + "\n")
	tiPassword := blurredStyle.Render(l.TIPassword.View())
	b.WriteString(tiPassword + "\n")

	form := containerStyle.Render(b.String())

	return lipgloss.Place(
		l.width, l.height,
		lipgloss.Center, lipgloss.Center,
		form)
}
